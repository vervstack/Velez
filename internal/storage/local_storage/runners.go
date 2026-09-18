package local_storage

import (
	"context"
	"sort"
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
	// runnerSecretScope / runnerSecretKey mirror runneraas.runnerSecretScope /
	// runneraas.runnerSecretKey - the scope+key half of the domain.SecretRef
	// the runneraas service stores a caller's access token under. Owner is
	// the instance name. Duplicated rather than shared: the runneraas
	// service package sits a layer above storage and must not be imported
	// here.
	runnerSecretScope = "runneraas"
	runnerSecretKey   = "access_token"
)

// dockerRunners is the single-node/dev storage.RunnersStorage: there is no
// velez.runners table, so a running container labelled
// labels.RunnerInstanceLabel is the system of record. Simpler than
// dockerPgInstances (pg_instances.go): provider/scope/target/labels are all
// container labels rather than env vars (see
// internal/domain/labels.RunnerInstanceLabel's doc comment), so no
// ContainerInspect call is needed - ListContainers' own container.Summary
// already carries everything. A small in-memory overlay covers the window
// between UpsertRunner and the deploy watcher actually creating that
// container (mirrors dockerPgInstances.pending).
type dockerRunners struct {
	docker node_clients.Docker

	mu      sync.Mutex
	pending map[int64]domain.Runner
}

func newRunnersStorage(docker node_clients.Docker) *dockerRunners {
	return &dockerRunners{
		docker:  docker,
		pending: make(map[int64]domain.Runner),
	}
}

// UpsertRunner records the row in the pending overlay and echoes it back -
// the container the deploy watcher is about to create is the durable copy.
func (d *dockerRunners) UpsertRunner(_ context.Context, req domain.UpsertRunnerReq) (domain.Runner, error) {
	now := time.Now()

	d.mu.Lock()
	defer d.mu.Unlock()

	existing, ok := d.pending[req.ServiceID]

	createdAt := now
	if ok {
		createdAt = existing.CreatedAt
	}

	runner := domain.Runner{
		ServiceID: req.ServiceID,
		Provider:  req.Provider,
		Scope:     req.Scope,
		Target:    req.Target,
		Labels:    req.Labels,
		SecretRef: req.SecretRef,
		BaseUrl:   req.BaseUrl,
		CreatedAt: createdAt,
		UpdatedAt: now,
	}

	d.pending[req.ServiceID] = runner

	return runner, nil
}

func (d *dockerRunners) GetRunnerByServiceID(ctx context.Context, serviceID int64) (domain.Runner, error) {
	runners, err := d.listFromContainers(ctx)
	if err != nil {
		return domain.Runner{}, err
	}

	for _, runner := range runners {
		if runner.ServiceID == serviceID {
			return runner, nil
		}
	}

	d.mu.Lock()

	pending, ok := d.pending[serviceID]

	d.mu.Unlock()

	if ok {
		return pending, nil
	}

	return domain.Runner{}, rerrors.Wrap(user_errors.ErrStorageNotFound)
}

func (d *dockerRunners) ListRunners(ctx context.Context) ([]domain.Runner, error) {
	runners, err := d.listFromContainers(ctx)
	if err != nil {
		return nil, err
	}

	live := make(map[int64]struct{}, len(runners))
	for _, runner := range runners {
		live[runner.ServiceID] = struct{}{}
	}

	d.mu.Lock()

	for id, pending := range d.pending {
		if _, ok := live[id]; ok {
			continue
		}

		runners = append(runners, pending)
	}

	d.mu.Unlock()

	sort.Slice(runners, func(i, j int) bool {
		return runners[i].ServiceID < runners[j].ServiceID
	})

	return runners, nil
}

// DeleteRunner only clears the pending overlay - dropping the container
// (runneraas.DropRunner calls VervServicesService.Remove first) is what
// actually removes a live instance. A missing overlay entry is not an error:
// the common case is deleting an instance whose container already exists.
func (d *dockerRunners) DeleteRunner(_ context.Context, serviceID int64) error {
	d.mu.Lock()
	delete(d.pending, serviceID)
	d.mu.Unlock()

	return nil
}

// listFromContainers rebuilds every runner's row from its labelled
// container's own labels - provider/scope/target/labels need no
// ContainerInspect call, unlike dockerPgInstances.listFromContainers, which
// has to inspect for env vars.
func (d *dockerRunners) listFromContainers(ctx context.Context) ([]domain.Runner, error) {
	listReq := &pb.ListSmerds_Request{
		Label: map[string]string{labels.RunnerInstanceLabel: boolLabelValue},
	}

	containers, err := d.docker.ListContainers(ctx, listReq, allEnvironments)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing runner containers")
	}

	out := make([]domain.Runner, 0, len(containers))

	for _, c := range containers {
		name := c.Labels[labels.VervServiceLabel]
		if name == "" && len(c.Names) != 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		if name == "" {
			continue
		}

		created := time.Unix(c.Created, 0)

		secretRef := domain.SecretRef{Scope: runnerSecretScope, Owner: name, Key: runnerSecretKey}

		runner := domain.Runner{
			ServiceID: serviceIDFromName(name),
			Provider:  c.Labels[labels.RunnerProviderLabel],
			Scope:     c.Labels[labels.RunnerScopeLabel],
			Target:    c.Labels[labels.RunnerTargetLabel],
			Labels:    splitRunnerLabels(c.Labels[labels.RunnerLabelsLabel]),
			SecretRef: secretRef.String(),
			BaseUrl:   c.Labels[labels.RunnerBaseUrlLabel],
			CreatedAt: created,
			UpdatedAt: created,
		}

		out = append(out, runner)
	}

	return out, nil
}

// splitRunnerLabels is the inverse of strings.Join(labels, ",") - an empty
// value means no labels were set, not one empty-string label.
func splitRunnerLabels(value string) []string {
	if value == "" {
		return nil
	}

	return strings.Split(value, ",")
}
