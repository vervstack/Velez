package local_storage

import (
	"context"
	"database/sql"
	"sort"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/container_derived"
	"go.vervstack.ru/Velez/internal/storage/registries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	// registryContainerPort is the port the registry listens on inside its container.
	registryContainerPort = 5000
)

// ContainerAddressResolver turns an inspected container into the host:port
// Velez dials to reach one of its ports.
type ContainerAddressResolver interface {
	Address(info container.InspectResponse, containerPort int) (string, error)
}

// ContainerAddressBinder is implemented by storages that resolve container
// addresses on read. The resolver exists only after the node's container
// runtime is wired, which is later than the storage itself, so it is bound late.
type ContainerAddressBinder interface {
	BindContainerAddressResolver(resolver ContainerAddressResolver)
}

// dockerRegistries is the single-node/dev storage.RegistriesStorage: the
// seeded Docker Hub row plus one row per registryaas container, rebuilt from
// the container on every read - including its url, which is resolved against
// the live container with the shared-daemon addressing. Registries on a
// dedicated-daemon environment are not listed, as with dockerRegistryInstances.
// The pending overlay only bridges the window before the container exists.
type dockerRegistries struct {
	docker node_clients.Docker

	mu       sync.Mutex
	resolver ContainerAddressResolver
	pending  map[string]domain.Registry
}

func newRegistriesStorage(docker node_clients.Docker) *dockerRegistries {
	return &dockerRegistries{
		docker:  docker,
		pending: make(map[string]domain.Registry),
	}
}

func (d *dockerRegistries) BindContainerAddressResolver(resolver ContainerAddressResolver) {
	d.mu.Lock()

	d.resolver = resolver
	d.mu.Unlock()
}

func (d *dockerRegistries) ListRegistries(ctx context.Context) ([]domain.Registry, error) {
	derived, err := d.listFromContainers(ctx)
	if err != nil {
		return nil, err
	}

	live := make(map[string]struct{}, len(derived))
	for _, reg := range derived {
		live[reg.Name] = struct{}{}
	}

	out := make([]domain.Registry, 0, len(derived)+1)

	out = append(out, registries.DockerHub())
	out = append(out, derived...)

	d.mu.Lock()

	for name, pending := range d.pending {
		_, isLive := live[name]
		if isLive {
			continue
		}

		out = append(out, pending)
	}

	d.mu.Unlock()

	sort.Slice(out, func(i, j int) bool {
		return out[i].Id < out[j].Id
	})

	return out, nil
}

func (d *dockerRegistries) GetRegistryByID(ctx context.Context, id int64) (domain.Registry, error) {
	all, err := d.ListRegistries(ctx)
	if err != nil {
		return domain.Registry{}, err
	}

	for _, reg := range all {
		if reg.Id == id {
			return reg, nil
		}
	}

	return domain.Registry{}, rerrors.Wrap(user_errors.ErrStorageNotFound)
}

func (d *dockerRegistries) CreateRegistry(_ context.Context, _ domain.CreateRegistryReq) (domain.Registry, error) {
	return domain.Registry{}, rerrors.Wrap(user_errors.ErrRequiresStatefullMode)
}

func (d *dockerRegistries) UpdateRegistry(_ context.Context, _ domain.UpdateRegistryReq) (domain.Registry, error) {
	return domain.Registry{}, rerrors.Wrap(user_errors.ErrRequiresStatefullMode)
}

func (d *dockerRegistries) DeleteRegistry(_ context.Context, _ int64) error {
	return rerrors.Wrap(user_errors.ErrRequiresStatefullMode)
}

func (d *dockerRegistries) ClearDefaultRegistry(_ context.Context) error {
	return rerrors.Wrap(user_errors.ErrRequiresStatefullMode)
}

func (d *dockerRegistries) WithTx(_ *sql.Tx) storage.RegistriesStorage {
	return d
}

// UpsertBuiltinRegistry only records the pending overlay - the container the
// create job just started is the durable copy and is re-read on every list.
func (d *dockerRegistries) UpsertBuiltinRegistry(
	_ context.Context, req domain.CreateRegistryReq,
) (domain.Registry, error) {
	now := time.Now()

	d.mu.Lock()
	defer d.mu.Unlock()

	createdAt := now

	existing, ok := d.pending[req.Name]
	if ok {
		createdAt = existing.CreatedAt
	}

	reg := domain.Registry{
		Id:        registryIdFromName(req.Name),
		Name:      req.Name,
		Type:      req.Type,
		Url:       req.Url,
		Username:  req.Username,
		Secret:    req.Secret,
		IsDefault: req.IsDefault,
		CreatedAt: createdAt,
		UpdatedAt: now,
	}

	d.pending[req.Name] = reg

	return reg, nil
}

// DeleteBuiltinRegistry only clears the pending overlay - dropping the
// container is what removes a live row.
func (d *dockerRegistries) DeleteBuiltinRegistry(_ context.Context, name string) error {
	d.mu.Lock()
	delete(d.pending, name)
	d.mu.Unlock()

	return nil
}

func (d *dockerRegistries) listFromContainers(ctx context.Context) ([]domain.Registry, error) {
	listReq := &pb.ListSmerds_Request{
		Label: map[string]string{labels.RegistryaasInstanceLabel: boolLabelValue},
	}

	containers, err := d.docker.ListContainers(ctx, listReq, allEnvironments)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing registryaas containers")
	}

	out := make([]domain.Registry, 0, len(containers))

	for _, c := range containers {
		name := container_derived.ServiceName(c.Labels, c.Names)
		if name == "" {
			continue
		}

		url, urlErr := d.registryUrl(ctx, c.ID)
		if urlErr != nil {
			return nil, urlErr
		}

		instance := container_derived.RegistryInstance(name, time.Unix(c.Created, 0), c.Labels)

		reg := domain.Registry{
			Id:        registryIdFromName(name),
			Name:      name,
			Type:      domain.RegistryTypeGenericV2,
			Url:       url,
			Username:  instance.Username,
			Secret:    instance.SecretRef,
			CreatedAt: instance.CreatedAt,
			UpdatedAt: instance.UpdatedAt,
		}

		out = append(out, reg)
	}

	return out, nil
}

// registryUrl is empty for a container whose address cannot be resolved
// right now (stopped, port unpublished), so one down registry does not hide
// the others from a list.
func (d *dockerRegistries) registryUrl(ctx context.Context, containerId string) (string, error) {
	d.mu.Lock()

	resolver := d.resolver
	d.mu.Unlock()

	if resolver == nil {
		return "", rerrors.Wrap(errAddressResolverNotBound)
	}

	info, err := d.docker.Client().ContainerInspect(ctx, containerId)
	if err != nil {
		return "", rerrors.Wrap(err, "error inspecting registry container")
	}

	address, err := resolver.Address(info, registryContainerPort)
	if err != nil {
		log.Ctx(ctx).Warn().
			Str("container", info.Name).
			Err(err).
			Msg("registry address is not resolvable")

		return "", nil
	}

	return "http://" + address, nil
}

// registryIdFromName is the name-derived id of a container-backed registry; it
// never equals the seeded Docker Hub id.
func registryIdFromName(name string) int64 {
	id := serviceIDFromName(name)
	if id == registries.DockerHubRegistryId {
		return id + 1
	}

	return id
}
