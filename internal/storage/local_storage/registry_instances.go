package local_storage

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	// registryaasSecretScope / registryaasSecretKey mirror
	// registryaas.registrySecretScope / registryaas.registrySecretKey - the
	// scope+key half of the domain.SecretRef the registryaas service stores
	// a generated password under. Owner is the instance name. Duplicated
	// rather than shared: the registryaas service package sits a layer above
	// storage and must not be imported here.
	registryaasSecretScope = "registryaas"
	registryaasSecretKey   = "password"

	// registryaasDefaultPort is the fallback used only when a container
	// predates labels.RegistryaasPortLabel (upgrade from an older Velez) -
	// the builtin registry descriptor's container-internal port, not a
	// usable host port, but the best guess available without it.
	registryaasDefaultPort = 5000
)

// dockerRegistryInstances is the single-node/dev storage.
// RegistryInstancesStorage: there is no velez.registry_instances table, so a
// running container labelled labels.RegistryaasInstanceLabel is the system
// of record. Credentials live in an htpasswd file rather than env vars, so -
// unlike dockerPgInstances - username and ui_port are read back from the
// container's own labels rather than an inspect-for-env call, mirroring
// dockerRunners. A small in-memory overlay covers the window between
// UpsertRegistryInstance and the deploy watcher actually creating that
// container (mirrors dockerPgInstances.pending).
type dockerRegistryInstances struct {
	docker node_clients.Docker

	mu      sync.Mutex
	pending map[int64]domain.RegistryInstance
}

func newRegistryInstancesStorage(docker node_clients.Docker) *dockerRegistryInstances {
	return &dockerRegistryInstances{
		docker:  docker,
		pending: make(map[int64]domain.RegistryInstance),
	}
}

// UpsertRegistryInstance records the row in the pending overlay and echoes it
// back - the container the deploy watcher is about to create is the durable
// copy.
func (d *dockerRegistryInstances) UpsertRegistryInstance(
	_ context.Context, req domain.UpsertRegistryInstanceReq,
) (domain.RegistryInstance, error) {
	now := time.Now()

	d.mu.Lock()
	defer d.mu.Unlock()

	existing, ok := d.pending[req.ServiceId]

	createdAt := now
	if ok {
		createdAt = existing.CreatedAt
	}

	instance := domain.RegistryInstance{
		ServiceId: req.ServiceId,
		Port:      req.Port,
		UiPort:    req.UiPort,
		Username:  req.Username,
		SecretRef: req.SecretRef,
		CreatedAt: createdAt,
		UpdatedAt: now,
	}

	d.pending[req.ServiceId] = instance

	return instance, nil
}

func (d *dockerRegistryInstances) GetRegistryInstanceByServiceID(
	ctx context.Context, serviceID int64,
) (domain.RegistryInstance, error) {
	instances, err := d.listFromContainers(ctx)
	if err != nil {
		return domain.RegistryInstance{}, err
	}

	for _, instance := range instances {
		if instance.ServiceId == serviceID {
			return instance, nil
		}
	}

	d.mu.Lock()

	pending, ok := d.pending[serviceID]

	d.mu.Unlock()

	if ok {
		return pending, nil
	}

	return domain.RegistryInstance{}, rerrors.Wrap(user_errors.ErrStorageNotFound)
}

func (d *dockerRegistryInstances) ListRegistryInstances(ctx context.Context) ([]domain.RegistryInstance, error) {
	instances, err := d.listFromContainers(ctx)
	if err != nil {
		return nil, err
	}

	live := make(map[int64]struct{}, len(instances))
	for _, instance := range instances {
		live[instance.ServiceId] = struct{}{}
	}

	d.mu.Lock()

	for id, pending := range d.pending {
		if _, ok := live[id]; ok {
			continue
		}

		instances = append(instances, pending)
	}

	d.mu.Unlock()

	sort.Slice(instances, func(i, j int) bool {
		return instances[i].ServiceId < instances[j].ServiceId
	})

	return instances, nil
}

// DeleteRegistryInstance only clears the pending overlay - dropping the
// container (registryaas.DropRegistryInstance calls VervServicesService.Remove
// first) is what actually removes a live instance. A missing overlay entry is
// not an error: the common case is deleting an instance whose container
// already exists.
func (d *dockerRegistryInstances) DeleteRegistryInstance(_ context.Context, serviceID int64) error {
	d.mu.Lock()
	delete(d.pending, serviceID)
	d.mu.Unlock()

	return nil
}

// listFromContainers rebuilds every instance's row from its labelled
// container's own labels - username/ui_port need no ContainerInspect call,
// unlike dockerPgInstances.listFromContainers, because credentials live in an
// htpasswd file rather than env vars.
func (d *dockerRegistryInstances) listFromContainers(ctx context.Context) ([]domain.RegistryInstance, error) {
	listReq := &pb.ListSmerds_Request{
		Label: map[string]string{labels.RegistryaasInstanceLabel: boolLabelValue},
	}

	containers, err := d.docker.ListContainers(ctx, listReq, allEnvironments)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing registryaas containers")
	}

	out := make([]domain.RegistryInstance, 0, len(containers))

	for _, c := range containers {
		name := c.Labels[labels.VervServiceLabel]
		if name == "" && len(c.Names) != 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		if name == "" {
			continue
		}

		created := time.Unix(c.Created, 0)

		secretRef := domain.SecretRef{Scope: registryaasSecretScope, Owner: name, Key: registryaasSecretKey}

		instance := domain.RegistryInstance{
			ServiceId: serviceIDFromName(name),
			Port:      portFromLabel(c.Labels[labels.RegistryaasPortLabel]),
			UiPort:    uiPortFromLabel(c.Labels[labels.RegistryaasUiPortLabel]),
			Username:  c.Labels[labels.RegistryaasUsernameLabel],
			SecretRef: secretRef.String(),
			CreatedAt: created,
			UpdatedAt: created,
		}

		out = append(out, instance)
	}

	return out, nil
}

// uiPortFromLabel parses labels.RegistryaasUiPortLabel's value, falling back
// to the 0 "not provisioned yet" sentinel (mirrors the backfill migration's
// ui_port = 0 convention) on an empty or malformed label.
func uiPortFromLabel(value string) int32 {
	port, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0
	}

	return int32(port) //nolint:gosec
}

// portFromLabel parses labels.RegistryaasPortLabel's value - the instance's
// resolved host-exposed port, set on the registry container at deploy time
// (deployRegistryInstanceJob). Falls back to registryaasDefaultPort (the
// container-internal port, not a real host port) only for a container
// created before this label existed.
func portFromLabel(value string) int32 {
	port, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return registryaasDefaultPort
	}

	return int32(port) //nolint:gosec
}
