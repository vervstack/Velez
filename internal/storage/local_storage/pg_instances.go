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

// dockerPgInstances is the single-node/dev storage.PgInstancesStorage: there
// is no velez.pg_instances table, so a running container labelled
// labels.PgaasInstanceLabel is the system of record. A small in-memory
// overlay covers the window between CreatePgInstance and the deploy watcher
// actually creating that container (mirrors dockerServices.upserted).
type dockerPgInstances struct {
	docker node_clients.Docker

	mu      sync.Mutex
	pending map[int64]domain.PgInstance
}

func newPgInstancesStorage(docker node_clients.Docker) *dockerPgInstances {
	return &dockerPgInstances{
		docker:  docker,
		pending: make(map[int64]domain.PgInstance),
	}
}

// UpsertPgInstance records the row in the pending overlay and echoes it back -
// the container the deploy watcher is about to create is the durable copy.
func (d *dockerPgInstances) UpsertPgInstance(
	_ context.Context, req domain.UpsertPgInstanceReq,
) (domain.PgInstance, error) {
	now := time.Now()

	d.mu.Lock()
	defer d.mu.Unlock()

	existing, ok := d.pending[req.ServiceId]

	createdAt := now
	if ok {
		createdAt = existing.CreatedAt
	}

	instance := domain.PgInstance{
		ServiceId: req.ServiceId,
		DbName:    req.DbName,
		Username:  req.Username,
		SecretRef: req.SecretRef,
		Port:      req.Port,
		CreatedAt: createdAt,
		UpdatedAt: now,
	}

	d.pending[req.ServiceId] = instance

	return instance, nil
}

func (d *dockerPgInstances) GetPgInstanceByServiceID(
	ctx context.Context, serviceID int64,
) (domain.PgInstance, error) {
	instances, err := d.listFromContainers(ctx)
	if err != nil {
		return domain.PgInstance{}, err
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

	return domain.PgInstance{}, rerrors.Wrap(user_errors.ErrStorageNotFound)
}

func (d *dockerPgInstances) ListPgInstances(ctx context.Context) ([]domain.PgInstance, error) {
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

// DeletePgInstance only clears the pending overlay - dropping the container
// (pgaas.DropPgInstance calls VervServicesService.Remove first) is what
// actually removes a live instance. A missing overlay entry is not an error:
// the common case is deleting an instance whose container already exists.
func (d *dockerPgInstances) DeletePgInstance(_ context.Context, serviceID int64) error {
	d.mu.Lock()
	delete(d.pending, serviceID)
	d.mu.Unlock()

	return nil
}

// listFromContainers rebuilds every instance's row from its labelled
// container. It inspects each container for the POSTGRES_* env pgaas wrote at
// deploy time - container.Summary carries labels but not env.
func (d *dockerPgInstances) listFromContainers(ctx context.Context) ([]domain.PgInstance, error) {
	listReq := &pb.ListSmerds_Request{
		Label: map[string]string{labels.PgaasInstanceLabel: boolLabelValue},
	}

	containers, err := d.docker.ListContainers(ctx, listReq, allEnvironments)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing pgaas containers")
	}

	out := make([]domain.PgInstance, 0, len(containers))

	for _, c := range containers {
		name := container_derived.ServiceName(c.Labels, c.Names)
		if name == "" {
			continue
		}

		info, inspectErr := d.docker.Client().ContainerInspect(ctx, c.ID)
		if inspectErr != nil {
			return nil, rerrors.Wrap(inspectErr, "error inspecting pgaas container")
		}

		instance := container_derived.PgInstance(name, time.Unix(c.Created, 0), info.Config.Env)

		instance.ServiceId = serviceIDFromName(name)

		out = append(out, instance)
	}

	return out, nil
}
