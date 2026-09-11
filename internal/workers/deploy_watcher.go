package workers

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
	"golang.org/x/sync/errgroup"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	// taskWatchTimeout bounds how long one scheduled deployment blocks waiting
	// for its create_smerd/upgrade_smerd task to reach a terminal status. Same
	// safety-net role as velez_api_impl's upgradeSmerdWatchTimeout, and sized
	// to it: upgrade_smerd is the slower of the two actions this worker runs.
	taskWatchTimeout = 120 * time.Second
)

// taskRunner is the slice of jobs.Engine this worker needs: enqueue a task
// and follow it to a terminal status.
type taskRunner interface {
	Enqueue(ctx context.Context, entityID, action string, initialContext any) (tasks_queries.VelezTask, error)
	Watch(ctx context.Context, entityID, action string) <-chan tasks_queries.VelezTask
}

// deploymentsStorageResolver yields the deployments storage of whatever
// backend is currently live, re-resolved per call so an enable_statefull
// swap is observed without a restart. Mirrors
// verv_services.VervService.environments() and the taskRunner interface above.
type deploymentsStorageResolver interface {
	Deployments() storage.DeploymentsStorage
}

type deployWatcher struct {
	// jobsEngine replaces the deleted internal/pipelines.Pipeliner: scheduled
	// deployments and upgrades are enqueued as durable tasks and awaited
	// synchronously, mirroring velez_api_impl's CreateSmerd/UpgradeSmerd
	// facades. The environment no longer has to be resolved into a Docker
	// suffix here - create_smerd's own jobs re-resolve it from the persisted
	// request at run time.
	jobsEngine  taskRunner
	dataStorage deploymentsStorageResolver
	// runtimes resolves a deployment's environment name into the
	// ContainerRuntime serving it, so the liveness check and the deletion
	// below stay scoped to the right environment instead of hitting the
	// daemon through an unsuffixed Docker client.
	runtimes container_runtime.RuntimeResolver

	nodeId int64

	starter  sync.Once
	stopOnce sync.Once
	ticker   *time.Ticker
	done     chan struct{}
}

func NewDeployWatcher(
	services service.Services,
	jobsEngine jobs.Engine,
	clusterClients cluster_clients.ClusterClients,
	runtimes container_runtime.RuntimeResolver,

	interval time.Duration,
) Worker {
	return &deployWatcher{
		jobsEngine:  jobsEngine,
		dataStorage: clusterClients.StateManager(),
		runtimes:    runtimes,

		nodeId: 1,

		starter: sync.Once{},
		ticker:  time.NewTicker(interval),
		done:    make(chan struct{}),
	}
}

func (d *deployWatcher) Start(ctx context.Context) {
	d.starter.Do(func() {
		for {
			select {
			case <-d.done:
				return
			case <-d.ticker.C:
				list, err := d.listDeployments(ctx)
				if err != nil {
					if !rerrors.Is(err, user_errors.ErrServiceIsDisabled) {
						log.Error().Err(err).Msg("error listing deployments in deploy watcher")

						// TODO make it fail only when state is not available
						// retry via api handle
						return
					}

					continue
				}

				g, errCtx := errgroup.WithContext(ctx)

				g.Go(func() error { return d.processScheduledBatch(errCtx, list.scheduled) })
				g.Go(func() error { return d.syncRunningBatch(errCtx, list.active) })
				g.Go(func() error { return d.deleteBatch(errCtx, list.scheduledDeletion) })

				err = g.Wait()
				if err != nil {
					log.Error().Err(err).Msg("error running deploy watcher")

					continue
				}
			}
		}
	})
}

func (d *deployWatcher) Stop() error {
	d.stopOnce.Do(func() {
		d.ticker.Stop()
		close(d.done)
	})

	return nil
}

// deployments re-resolves the live deployments storage per call so an
// enable_statefull swap of the cluster state manager is observed without a
// restart. Mirrors verv_services.VervService.environments().
func (d *deployWatcher) deployments() storage.DeploymentsStorage {
	return d.dataStorage.Deployments()
}

type deploymentsList struct {
	scheduled         []domain.Deployment
	active            []domain.Deployment
	scheduledDeletion []domain.Deployment
}

func (d *deployWatcher) listDeployments(ctx context.Context) (deploymentsList, error) {
	listReq := domain.ListDeploymentsReq{
		NodeIds: []int64{d.nodeId},
		NotStatus: []deployments_queries.VelezDeploymentStatus{
			deployments_queries.VelezDeploymentStatusDELETED,
			deployments_queries.VelezDeploymentStatusFAILED,
		},
	}

	deployments, err := d.deployments().List(ctx, listReq)
	if err != nil {
		return deploymentsList{}, rerrors.Wrap(err, "error listing deployments")
	}

	var list deploymentsList

	for _, dep := range deployments {
		switch dep.Status {
		case deployments_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT,
			deployments_queries.VelezDeploymentStatusSCHEDULEDUPGRADE:
			list.scheduled = append(list.scheduled, dep)
		case deployments_queries.VelezDeploymentStatusRUNNING:
			list.active = append(list.active, dep)
		case deployments_queries.VelezDeploymentStatusSCHEDULEDDELETION:
			list.scheduledDeletion = append(list.scheduledDeletion, dep)

		default:
			log.Error().Any("status", dep.Status).Msg("unknown deployment status in deploy watcher")
		}
	}

	return list, nil
}

func (d *deployWatcher) processScheduledBatch(ctx context.Context, scheduled []domain.Deployment) error {
	for _, dep := range scheduled {
		//	Step 1 - define what's needs to  be done - upgrade or new deployment
		//nolint:exhaustive
		switch dep.Status {
		case deployments_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT:
			err := d.deploy(ctx, dep)
			if err != nil {
				return rerrors.Wrap(err)
			}
		case deployments_queries.VelezDeploymentStatusSCHEDULEDDELETION:
		case deployments_queries.VelezDeploymentStatusSCHEDULEDUPGRADE:
			err := d.upgrade(ctx, dep)
			if err != nil {
				return rerrors.Wrap(err)
			}
		}
	}

	return nil
}

// deploy launches a scheduled deployment through the create_smerd task.
// Unlike the pipeliner it replaces, the environment name is not resolved into
// a Docker suffix here: the persisted request carries the name, and
// create_smerd's container-creation job resolves it against the live
// environments storage when it runs.
func (d *deployWatcher) deploy(ctx context.Context, dep domain.Deployment) error {
	smerdReq, err := d.specRequest(ctx, dep)
	if err != nil {
		return rerrors.Wrap(err)
	}

	updateStatusParams := deployments_queries.UpdateDeploymentStatusParams{
		Status: deployments_queries.VelezDeploymentStatusRUNNING,
		ID:     dep.Id,
	}

	initialContext := &velez_api.CreateSmerdTaskPayload{}
	initialContext.SetRequest(smerdReq)

	// velez.tasks is UNIQUE (entity_id, action): scope the entity id by
	// environment so the same service name deployed into two environments
	// doesn't dedup onto one task. This worker deliberately never resolves the
	// environment name into a Docker suffix (create_smerd's own jobs do that
	// at run time), so the stored environment name is what scopes the key -
	// stable per environment and empty for the default one, which keeps the
	// historical bare-name id. TODO(#127): fold in the resolved suffix if this
	// worker ever gains access to environments storage.
	entityID := jobs.SmerdEntityID(smerdReq.GetEnvironment(), smerdReq.GetName())

	err = d.runTask(ctx, entityID, jobs.CreateSmerdAction, initialContext)
	if err != nil {
		log.Error().Err(rerrors.Wrap(err, "")).Msg("error deploying smerd")

		updateStatusParams.Status = deployments_queries.VelezDeploymentStatusFAILED
	}

	err = d.deployments().UpdateDeploymentStatus(ctx, updateStatusParams)
	if err != nil {
		return rerrors.Wrap(err, "UpdateDeploymentStatus")
	}

	return nil
}

// upgrade upgrades a running deployment through the upgrade_smerd task. The
// stored specification's environment is now carried into the request - the
// pipeliner path read it for deployments but silently dropped it for
// upgrades, which made every upgrade implicitly target the default
// environment.
func (d *deployWatcher) upgrade(ctx context.Context, dep domain.Deployment) error {
	smerdReq, err := d.specRequest(ctx, dep)
	if err != nil {
		return rerrors.Wrap(err)
	}

	updateStatusParams := deployments_queries.UpdateDeploymentStatusParams{
		Status: deployments_queries.VelezDeploymentStatusRUNNING,
		ID:     dep.Id,
	}

	upgradeReq := &velez_api.UpgradeSmerd_Request{
		Name:        smerdReq.GetName(),
		Image:       smerdReq.GetImageName(),
		Environment: smerdReq.GetEnvironment(),
	}

	initialContext := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: upgradeReq,
	}

	// Scope the entity id by environment - see the note in deploy(). TODO(#127).
	entityID := jobs.SmerdEntityID(upgradeReq.GetEnvironment(), upgradeReq.GetName())

	err = d.runTask(ctx, entityID, jobs.UpgradeSmerdAction, initialContext)
	if err != nil {
		log.Error().Err(rerrors.Wrap(err, "")).Msg("error upgrading smerd")

		updateStatusParams.Status = deployments_queries.VelezDeploymentStatusFAILED
	}

	err = d.deployments().UpdateDeploymentStatus(ctx, updateStatusParams)
	if err != nil {
		return rerrors.Wrap(err, "UpdateDeploymentStatus")
	}

	return nil
}

func (d *deployWatcher) syncRunningBatch(ctx context.Context, active []domain.Deployment) error {
	for _, dep := range active {
		smerdReq, err := d.specRequest(ctx, dep)
		if err != nil {
			return rerrors.Wrap(err, "error getting spec for running deployment")
		}

		runtime, err := d.runtimes.Runtime(ctx, smerdReq.GetEnvironment())
		if err != nil {
			log.Error().Err(err).Str("container", smerdReq.GetName()).Msg("error resolving container runtime")

			continue
		}

		running, _, err := runtime.IsContainerRunning(ctx, smerdReq.GetName())
		if err != nil {
			log.Error().Err(err).Str("container", smerdReq.GetName()).Msg("error inspecting container")

			continue
		}

		if running {
			continue
		}

		updateStatusParams := deployments_queries.UpdateDeploymentStatusParams{
			Status: deployments_queries.VelezDeploymentStatusFAILED,
			ID:     dep.Id,
		}

		err = d.deployments().UpdateDeploymentStatus(ctx, updateStatusParams)
		if err != nil {
			return rerrors.Wrap(err, "error marking deployment as failed")
		}
	}

	return nil
}

func (d *deployWatcher) deleteBatch(ctx context.Context, deletion []domain.Deployment) error {
	for _, dep := range deletion {
		smerdReq, err := d.specRequest(ctx, dep)
		if err != nil {
			return rerrors.Wrap(err, "error getting spec for deletion")
		}

		runtime, err := d.runtimes.Runtime(ctx, smerdReq.GetEnvironment())
		if err != nil {
			log.Error().Err(err).Str("container", smerdReq.GetName()).Msg("error resolving container runtime")

			continue
		}

		err = runtime.Remove(ctx, smerdReq.GetName())
		if err != nil {
			log.Error().Err(err).Str("container", smerdReq.GetName()).Msg("error removing container")

			continue
		}

		updateStatusParams := deployments_queries.UpdateDeploymentStatusParams{
			Status: deployments_queries.VelezDeploymentStatusDELETED,
			ID:     dep.Id,
		}

		err = d.deployments().UpdateDeploymentStatus(ctx, updateStatusParams)
		if err != nil {
			return rerrors.Wrap(err, "error marking deployment as deleted")
		}
	}

	return nil
}

// specRequest loads a deployment's persisted specification and decodes the
// CreateSmerd request embedded in it.
func (d *deployWatcher) specRequest(
	ctx context.Context, dep domain.Deployment,
) (*velez_api.CreateSmerd_Request, error) {
	spec, err := d.deployments().GetSpecificationById(ctx, dep.SpecId)
	if err != nil {
		return nil, rerrors.Wrap(err, "GetSpecificationById")
	}

	smerdReq := &velez_api.CreateSmerd_Request{}

	err = json.Unmarshal(spec.VervPayload.RawMessage, smerdReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error unmarshaling spec")
	}

	return smerdReq, nil
}

// runTask enqueues action for entityID and blocks until the task reaches a
// terminal status - the synchronous facade over the jobs engine that
// velez_api_impl's CreateSmerd/UpgradeSmerd RPCs also put in front of it,
// keeping this worker's "run it and report success or failure" contract
// unchanged from the pipeliner it replaced.
func (d *deployWatcher) runTask(ctx context.Context, entityID, action string, initialContext any) error {
	_, err := d.jobsEngine.Enqueue(ctx, entityID, action, initialContext)
	if err != nil {
		return rerrors.Wrapf(err, "error enqueuing %s task", action)
	}

	watchCtx, cancel := context.WithTimeout(ctx, taskWatchTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	for task := range d.jobsEngine.Watch(watchCtx, entityID, action) {
		finalTask = task
	}

	isDone := finalTask.Status == tasks_queries.VelezTaskStatusDONE
	isFailed := finalTask.Status == tasks_queries.VelezTaskStatusFAILED

	if !isDone && !isFailed && watchCtx.Err() != nil {
		return rerrors.Wrapf(
			watchCtx.Err(),
			"timed out waiting for %s task, last status: %q",
			action,
			finalTask.Status,
		)
	}

	if isFailed {
		return user_errors.New(finalTask.Error.String)
	}

	return nil
}
