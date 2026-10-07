package jobs

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
)

const (
	containerStateRunning = "running"
)

// DeploymentReconciler registers a fresh RUNNING deployment for every service
// whose Velez-labelled container is running but has no live deployment row:
// the deploy watcher marks a row FAILED while its container is down and never
// revives it, and containers that appear after the switch to statefull mode
// are never backfilled. FAILED rows stay as history.
type DeploymentReconciler struct {
	storageContainer *storage.Container
	runtimes         container_runtime.RuntimeResolver
}

func NewDeploymentReconciler(
	storageContainer *storage.Container, runtimes container_runtime.RuntimeResolver,
) *DeploymentReconciler {
	return &DeploymentReconciler{
		storageContainer: storageContainer,
		runtimes:         runtimes,
	}
}

// runningPrimary is the running primary container of a service, taken from a
// container listing without inspecting it.
type runningPrimary struct {
	environment string
	serviceName string
	summary     container.Summary
}

// environmentSummaries is one environment's container listing.
type environmentSummaries struct {
	environment string
	summaries   []container.Summary
}

func (r *DeploymentReconciler) Reconcile(ctx context.Context) error {
	if !r.storageContainer.IsStatefull() {
		return nil
	}

	environments, err := r.storageContainer.Environments().ListEnvironments(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error listing environments")
	}

	for _, environment := range environments {
		reconcileErr := r.reconcileEnvironment(ctx, environment.Name)
		if reconcileErr != nil {
			log.Ctx(ctx).Warn().
				Str("environment", environment.Name).
				Err(reconcileErr).
				Msg("error reconciling deployments of environment, skipping it")
		}
	}

	return nil
}

func (r *DeploymentReconciler) reconcileEnvironment(ctx context.Context, environment string) error {
	runtime, err := r.runtimes.Runtime(ctx, environment)
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	summaries, err := runtime.ListAllContainers(ctx, listAllContainersNoLimit)
	if err != nil {
		return rerrors.Wrap(err, "error listing containers")
	}

	listed := environmentSummaries{environment: environment, summaries: summaries}

	for _, primary := range runningPrimaries(listed) {
		reconcileErr := r.reconcileService(ctx, runtime, primary)
		if reconcileErr != nil {
			log.Ctx(ctx).Warn().
				Str("service", primary.serviceName).
				Str("environment", environment).
				Err(reconcileErr).
				Msg("error reconciling service deployment, skipping it")
		}
	}

	return nil
}

func (r *DeploymentReconciler) reconcileService(
	ctx context.Context, runtime container_runtime.ContainerRuntime, primary runningPrimary,
) error {
	listReq := domain.ListDeploymentsReq{
		ServiceName: primary.serviceName,
		NotStatus: []deployments_queries.VelezDeploymentStatus{
			deployments_queries.VelezDeploymentStatusFAILED,
			deployments_queries.VelezDeploymentStatusDELETED,
		},
	}

	live, err := r.storageContainer.Deployments().List(ctx, listReq)
	if err != nil {
		return rerrors.Wrap(err, "error listing service deployments")
	}

	if len(live) != 0 {
		return nil
	}

	info, isFound, err := runtime.Inspect(ctx, primary.summary.ID)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	if !isFound || info.Config == nil {
		return nil
	}

	svc := labeledService{
		name:        primary.serviceName,
		displayName: displayNameOrName(info.Config.Labels, primary.serviceName),
		primary:     labeledContainer{environment: primary.environment, info: info},
	}

	err = registerRunningDeployment(ctx, r.storageContainer, svc, uuid.New().String())
	if err != nil {
		return rerrors.Wrap(err)
	}

	log.Ctx(ctx).Info().
		Str("service", primary.serviceName).
		Str("environment", primary.environment).
		Msg("registered running container that had no live deployment")

	return nil
}

// runningPrimaries returns, per service in first-seen order, its primary
// container (the first running one, else the first one) when that primary is
// running. Sidecars and containers without VERV_SERVICE are ignored.
func runningPrimaries(listed environmentSummaries) []runningPrimary {
	var primaries []runningPrimary

	indexByName := make(map[string]int)

	for _, summary := range listed.summaries {
		if !isBackfillCandidate(summary.Labels) {
			continue
		}

		name := summary.Labels[labels.VervServiceLabel]

		idx, isKnown := indexByName[name]
		if !isKnown {
			indexByName[name] = len(primaries)
			primaries = append(primaries, runningPrimary{
				environment: listed.environment,
				serviceName: name,
				summary:     summary,
			})

			continue
		}

		isPrimaryRunning := primaries[idx].summary.State == containerStateRunning
		if !isPrimaryRunning && summary.State == containerStateRunning {
			primaries[idx].summary = summary
		}
	}

	var running []runningPrimary

	for _, primary := range primaries {
		if primary.summary.State == containerStateRunning {
			running = append(running, primary)
		}
	}

	return running
}
