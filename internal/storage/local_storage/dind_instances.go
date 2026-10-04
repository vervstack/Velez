package local_storage

import (
	"context"
	"sort"
	"sync"
	"time"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/container_derived"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// dockerDindInstances is the single-node/dev storage.DindInstancesStorage:
// there is no velez.dind_instances table, so a running container labelled
// labels.DindInstanceLabel is the system of record. A small in-memory overlay
// covers the window between UpsertDindInstance and the deploy watcher
// actually creating that container (mirrors dockerRegistryInstances.pending).
type dockerDindInstances struct {
	docker node_clients.Docker

	mu      sync.Mutex
	pending map[int64]domain.DindInstance
}

func newDindInstancesStorage(docker node_clients.Docker) *dockerDindInstances {
	return &dockerDindInstances{
		docker:  docker,
		pending: make(map[int64]domain.DindInstance),
	}
}

func (d *dockerDindInstances) UpsertDindInstance(
	_ context.Context, req domain.UpsertDindInstanceReq,
) (domain.DindInstance, error) {
	now := time.Now()

	d.mu.Lock()
	defer d.mu.Unlock()

	existing, ok := d.pending[req.ServiceId]

	createdAt := now
	if ok {
		createdAt = existing.CreatedAt
	}

	instance := domain.DindInstance{
		ServiceId:       req.ServiceId,
		IsSysboxEnabled: req.IsSysboxEnabled,
		CreatedAt:       createdAt,
		UpdatedAt:       now,
	}

	d.pending[req.ServiceId] = instance

	return instance, nil
}

func (d *dockerDindInstances) GetDindInstanceByServiceId(
	ctx context.Context, serviceId int64,
) (domain.DindInstance, error) {
	instances, err := d.listFromContainers(ctx)
	if err != nil {
		return domain.DindInstance{}, err
	}

	for _, instance := range instances {
		if instance.ServiceId == serviceId {
			return instance, nil
		}
	}

	d.mu.Lock()

	pending, ok := d.pending[serviceId]

	d.mu.Unlock()

	if ok {
		return pending, nil
	}

	return domain.DindInstance{}, rerrors.Wrap(user_errors.ErrStorageNotFound)
}

func (d *dockerDindInstances) ListDindInstances(ctx context.Context) ([]domain.DindInstance, error) {
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

// DeleteDindInstance only clears the pending overlay - dropping the container
// is what actually removes a live instance.
func (d *dockerDindInstances) DeleteDindInstance(_ context.Context, serviceId int64) error {
	d.mu.Lock()
	delete(d.pending, serviceId)
	d.mu.Unlock()

	return nil
}

func (d *dockerDindInstances) listFromContainers(ctx context.Context) ([]domain.DindInstance, error) {
	listReq := &pb.ListSmerds_Request{
		Label: map[string]string{labels.DindInstanceLabel: boolLabelValue},
	}

	containers, err := d.docker.ListContainers(ctx, listReq, allEnvironments)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing dind containers")
	}

	out := make([]domain.DindInstance, 0, len(containers))

	for _, c := range containers {
		name := container_derived.ServiceName(c.Labels, c.Names)
		if name == "" {
			continue
		}

		instance := container_derived.DindInstance(name, time.Unix(c.Created, 0), c.Labels)

		instance.ServiceId = serviceIDFromName(name)

		out = append(out, instance)
	}

	return out, nil
}
