package local_storage

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/gitlab_runner_config"
	"go.vervstack.ru/Velez/internal/storage/container_derived"
	"go.vervstack.ru/Velez/internal/user_errors"
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
		ServiceID:           req.ServiceID,
		Provider:            req.Provider,
		Scope:               req.Scope,
		Target:              req.Target,
		Labels:              req.Labels,
		SecretRef:           req.SecretRef,
		BaseUrl:             req.BaseUrl,
		DockerImage:         req.DockerImage,
		DockerSocketAddress: req.DockerSocketAddress,
		Concurrent:          req.Concurrent,
		DindServiceId:       req.DindServiceId,
		CreatedAt:           createdAt,
		UpdatedAt:           now,
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
		name := container_derived.ServiceName(c.Labels, c.Names)
		if name == "" {
			continue
		}

		runner := container_derived.Runner(name, time.Unix(c.Created, 0), c.Labels)

		runner.ServiceID = serviceIDFromName(name)

		dindName := c.Labels[labels.RunnerDindLabel]
		if dindName != "" {
			runner.DindServiceId = serviceIDFromName(dindName)
		}

		if container_derived.IsGitlabRunner(runner) {
			runner.Concurrent = d.readGitlabConcurrent(ctx, c.ID, name)
		}

		out = append(out, runner)
	}

	return out, nil
}

// readGitlabConcurrent returns the `concurrent` value the container's own
// config.toml currently holds - a running container is the system of record
// here, and Docker labels can't be edited after create. 0 (treated as 1) when
// the file is unreadable or the key is absent, e.g. before Register has run.
func (d *dockerRunners) readGitlabConcurrent(ctx context.Context, containerID, name string) int32 {
	config, err := dockerutils.ReadFromContainer(ctx, d.docker.Client(), containerID, gitlab_runner_config.ConfigPath)
	if err != nil {
		log.Ctx(ctx).Warn().
			Str("runner", name).
			Err(err).
			Msg("error reading gitlab-runner config.toml")

		return 0
	}

	concurrent, _ := gitlab_runner_config.Concurrent(config)

	return concurrent
}
