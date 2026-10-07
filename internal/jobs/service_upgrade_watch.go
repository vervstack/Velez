package jobs

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

// DeploymentsResolver yields the deployments storage of whatever backend is
// currently live, re-resolved per call so an enable_statefull swap is
// observed without a restart.
type DeploymentsResolver interface {
	Deployments() storage.DeploymentsStorage
}

// upgradeTasks is the slice of Engine the watcher reads.
type upgradeTasks interface {
	Latest(ctx context.Context, entityID, action string) (sql.Null[tasks_queries.VelezTask], error)
	ListJobs(ctx context.Context, task tasks_queries.VelezTask) ([]JobStatus, error)
}

// ServiceUpgradeSnapshot is one observation of a service's upgrade.
type ServiceUpgradeSnapshot struct {
	// Task is invalid while the upgrade is only scheduled.
	Task sql.Null[tasks_queries.VelezTask]
	Jobs []JobStatus
}

// ServiceUpgradeWatcher follows a service's upgrade from the moment the
// deployment row is scheduled until its upgrade_smerd task finishes.
type ServiceUpgradeWatcher struct {
	deployments  DeploymentsResolver
	tasks        upgradeTasks
	pollInterval time.Duration
}

func NewServiceUpgradeWatcher(deployments DeploymentsResolver, engine Engine) *ServiceUpgradeWatcher {
	return &ServiceUpgradeWatcher{
		deployments:  deployments,
		tasks:        engine,
		pollInterval: defaultWatchPollInterval,
	}
}

// Watch streams snapshots until the upgrade reaches a terminal status, then
// closes the channel. It closes without sending when nothing is in flight.
func (w *ServiceUpgradeWatcher) Watch(ctx context.Context, serviceName string) <-chan ServiceUpgradeSnapshot {
	ch := make(chan ServiceUpgradeSnapshot)

	go func() {
		defer close(ch)

		run := &serviceUpgradeRun{
			watcher:     w,
			serviceName: serviceName,
			out:         ch,
		}

		ticker := time.NewTicker(w.pollInterval)
		defer ticker.Stop()

		for {
			isFinished, err := run.poll(ctx)
			if err != nil {
				log.Error().Err(err).Str("service_name", serviceName).Msg("error watching service upgrade")

				return
			}

			if isFinished {
				return
			}

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

type serviceUpgradeRun struct {
	watcher     *ServiceUpgradeWatcher
	serviceName string
	out         chan<- ServiceUpgradeSnapshot

	// entityID outlives the scheduled row: the row disappears once the task is done.
	entityID        string
	scheduledId     int64
	isTaskSeen      bool
	isAnythingSent  bool
	lastSent        ServiceUpgradeSnapshot
	isLastSentValid bool
}

// poll reports isFinished once the stream has nothing left to say.
func (r *serviceUpgradeRun) poll(ctx context.Context) (isFinished bool, err error) {
	scheduled, hasScheduled, err := r.newestScheduled(ctx)
	if err != nil {
		return false, rerrors.Wrap(err)
	}

	if hasScheduled && scheduled.Id != r.scheduledId {
		r.entityID, err = r.entityIdOf(ctx, scheduled)
		if err != nil {
			return false, rerrors.Wrap(err)
		}

		r.scheduledId = scheduled.Id
	}

	var latest sql.Null[tasks_queries.VelezTask]

	if r.entityID != "" {
		latest, err = r.watcher.tasks.Latest(ctx, r.entityID, UpgradeSmerdAction)
		if err != nil {
			return false, rerrors.Wrap(err)
		}
	}

	isActive := latest.Valid && isTaskActive(latest.V.Status)
	isTerminal := latest.Valid && !isActive

	switch {
	case isActive:
		r.isTaskSeen = true

		return false, r.sendTask(ctx, latest.V)
	case hasScheduled && r.isTaskSeen && isTerminal:
		err = r.sendTask(ctx, latest.V)

		return true, err
	case hasScheduled:
		return false, r.send(ctx, ServiceUpgradeSnapshot{})
	case r.isAnythingSent && isTerminal:
		err = r.sendTask(ctx, latest.V)

		return true, err
	default:
		return true, nil
	}
}

func isTaskActive(status tasks_queries.VelezTaskStatus) bool {
	return status == tasks_queries.VelezTaskStatusPENDING || status == tasks_queries.VelezTaskStatusRUNNING
}

func (r *serviceUpgradeRun) newestScheduled(ctx context.Context) (domain.Deployment, bool, error) {
	listReq := domain.ListDeploymentsReq{
		ServiceName: r.serviceName,
		NotStatus: []deployments_queries.VelezDeploymentStatus{
			deployments_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT,
			deployments_queries.VelezDeploymentStatusSCHEDULEDDELETION,
			deployments_queries.VelezDeploymentStatusRUNNING,
			deployments_queries.VelezDeploymentStatusFAILED,
			deployments_queries.VelezDeploymentStatusDELETED,
		},
	}

	list, err := r.watcher.deployments.Deployments().List(ctx, listReq)
	if err != nil {
		return domain.Deployment{}, false, rerrors.Wrap(err, "error listing scheduled upgrades")
	}

	if len(list) == 0 {
		return domain.Deployment{}, false, nil
	}

	newest := slices.MaxFunc(list, func(a, b domain.Deployment) int {
		return cmp.Compare(a.Id, b.Id)
	})

	return newest, true, nil
}

// entityIdOf mirrors the deploy watcher: the spec's environment NAME, not the suffix (TODO #127).
func (r *serviceUpgradeRun) entityIdOf(ctx context.Context, dep domain.Deployment) (string, error) {
	spec, err := r.watcher.deployments.Deployments().GetSpecificationById(ctx, dep.SpecId)
	if err != nil {
		return "", rerrors.Wrap(err, "error reading upgrade specification")
	}

	smerdReq := &velez_api.CreateSmerd_Request{}

	err = json.Unmarshal(spec.VervPayload.RawMessage, smerdReq)
	if err != nil {
		return "", rerrors.Wrap(err, "error unmarshaling upgrade specification")
	}

	return SmerdEntityID(smerdReq.GetEnvironment(), smerdReq.GetName()), nil
}

func (r *serviceUpgradeRun) sendTask(ctx context.Context, task tasks_queries.VelezTask) error {
	jobStatuses, err := r.watcher.tasks.ListJobs(ctx, task)
	if err != nil {
		return rerrors.Wrap(err, "error listing upgrade jobs")
	}

	snapshot := ServiceUpgradeSnapshot{
		Task: sql.Null[tasks_queries.VelezTask]{V: task, Valid: true},
		Jobs: jobStatuses,
	}

	return r.send(ctx, snapshot)
}

// send skips a snapshot equal to the previous one; a cancelled ctx is not an error.
func (r *serviceUpgradeRun) send(ctx context.Context, snapshot ServiceUpgradeSnapshot) error {
	if r.isLastSentValid && isSameSnapshot(r.lastSent, snapshot) {
		return nil
	}

	select {
	case r.out <- snapshot:
		r.lastSent = snapshot
		r.isLastSentValid = true
		r.isAnythingSent = true
	case <-ctx.Done():
	}

	return nil
}

func isSameSnapshot(a, b ServiceUpgradeSnapshot) bool {
	if a.Task.Valid != b.Task.Valid {
		return false
	}

	if a.Task.Valid && (a.Task.V.ID != b.Task.V.ID || a.Task.V.Status != b.Task.V.Status) {
		return false
	}

	return slices.Equal(a.Jobs, b.Jobs)
}
