package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	"github.com/sqlc-dev/pqtype"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/gitlab_runner_config"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/container_derived"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	stepBackfillLabeledServices = "backfill_labeled_services"

	listAllContainersNoLimit = 0
)

// labeledContainer is a Velez-labelled container together with the environment
// whose runtime owns it. info.Name is the environment's virtual name.
type labeledContainer struct {
	environment string
	info        container.InspectResponse
}

func (c labeledContainer) serviceName() string {
	return c.info.Config.Labels[labels.VervServiceLabel]
}

func (c labeledContainer) containerName() string {
	if c.info.ContainerJSONBase == nil {
		return ""
	}

	return strings.TrimPrefix(c.info.Name, "/")
}

func (c labeledContainer) containerId() string {
	if c.info.ContainerJSONBase == nil {
		return ""
	}

	return c.info.ID
}

func (c labeledContainer) isRunning() bool {
	return c.info.ContainerJSONBase != nil && c.info.State != nil && c.info.State.Running
}

func (c labeledContainer) hasLabel(label string) bool {
	_, ok := c.info.Config.Labels[label]

	return ok
}

// labeledService is one VERV_SERVICE value with the containers carrying it.
// primary is the container the deployment spec is built from; pg, runner and
// registry are the containers carrying the matching AAS label, if any.
type labeledService struct {
	name        string
	displayName string
	primary     labeledContainer
	pg          *labeledContainer
	runner      *labeledContainer
	registry    *labeledContainer
}

// isBackfillCandidate reports whether a container with these labels is a
// service container: carries VERV_SERVICE and is not a sidecar.
func isBackfillCandidate(containerLabels map[string]string) bool {
	if containerLabels[labels.VervServiceLabel] == "" {
		return false
	}

	_, isSidecar := containerLabels[labels.Sidecar]

	return !isSidecar
}

// groupLabeledServices groups candidate containers by service name, keeping
// first-seen order. The primary container is the first running one, else the
// first one.
func groupLabeledServices(containers []labeledContainer) []labeledService {
	var services []labeledService

	indexByName := make(map[string]int)

	for _, c := range containers {
		if c.info.Config == nil || !isBackfillCandidate(c.info.Config.Labels) {
			continue
		}

		name := c.serviceName()

		idx, isKnown := indexByName[name]
		if !isKnown {
			idx = len(services)
			indexByName[name] = idx

			services = append(services, labeledService{
				name:        name,
				displayName: displayNameOrName(c.info.Config.Labels, name),
				primary:     c,
			})
		}

		svc := &services[idx]

		if !svc.primary.isRunning() && c.isRunning() {
			svc.primary = c
		}

		assignAasContainer(svc, c)
	}

	return services
}

func assignAasContainer(svc *labeledService, c labeledContainer) {
	if c.hasLabel(labels.PgaasInstanceLabel) && svc.pg == nil {
		svc.pg = &c
	}

	if c.hasLabel(labels.RunnerInstanceLabel) && svc.runner == nil {
		svc.runner = &c
	}

	if c.hasLabel(labels.RegistryaasInstanceLabel) && svc.registry == nil {
		svc.registry = &c
	}
}

func displayNameOrName(containerLabels map[string]string, name string) string {
	displayName := containerLabels[labels.DisplayNameLabel]
	if displayName == "" {
		return name
	}

	return displayName
}

// backfillLabeledServicesJob registers every Velez-labelled container that
// existed before the switch to statefull mode (and the AAS satellite rows
// single-node storage derived from them) in Postgres. It only reads
// containers, and skips whatever already has rows, so a resumed run is safe.
// storageContainer and secrets are the live ones: update_cluster_state has
// already swapped them to Postgres by the time this runs.
type backfillLabeledServicesJob struct {
	storageContainer *storage.Container
	runtimes         container_runtime.RuntimeResolver
	secrets          secrets.Store
	dockerAPI        client.APIClient
}

func (j *backfillLabeledServicesJob) Do(ctx context.Context) error {
	containers, err := j.collectLabeledContainers(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error collecting labeled containers")
	}

	for _, svc := range groupLabeledServices(containers) {
		err = j.backfillService(ctx, svc)
		if err != nil {
			return rerrors.Wrap(err, "error backfilling service "+svc.name)
		}
	}

	return nil
}

func (j *backfillLabeledServicesJob) collectLabeledContainers(ctx context.Context) ([]labeledContainer, error) {
	environments, err := j.storageContainer.Environments().ListEnvironments(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing environments")
	}

	var containers []labeledContainer

	for _, environment := range environments {
		found, listErr := j.listEnvironmentContainers(ctx, environment.Name)
		if listErr != nil {
			log.Ctx(ctx).Warn().
				Str("environment", environment.Name).
				Err(listErr).
				Msg("error listing containers of environment, skipping it in backfill")

			continue
		}

		containers = append(containers, found...)
	}

	return containers, nil
}

func (j *backfillLabeledServicesJob) listEnvironmentContainers(
	ctx context.Context, environment string,
) ([]labeledContainer, error) {
	runtime, err := j.runtimes.Runtime(ctx, environment)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving container runtime")
	}

	summaries, err := runtime.ListAllContainers(ctx, listAllContainersNoLimit)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	var containers []labeledContainer

	for _, summary := range summaries {
		if !isBackfillCandidate(summary.Labels) {
			continue
		}

		info, isFound, inspectErr := runtime.Inspect(ctx, summary.ID)
		if inspectErr != nil {
			return nil, rerrors.Wrap(inspectErr, "error inspecting container")
		}

		if !isFound || info.Config == nil {
			continue
		}

		containers = append(containers, labeledContainer{environment: environment, info: info})
	}

	return containers, nil
}

func (j *backfillLabeledServicesJob) backfillService(ctx context.Context, svc labeledService) error {
	listReq := domain.ListDeploymentsReq{ServiceName: svc.name}

	deployments, err := j.storageContainer.Deployments().List(ctx, listReq)
	if err != nil {
		return rerrors.Wrap(err, "error listing service deployments")
	}

	if len(deployments) == 0 {
		err = registerRunningDeployment(ctx, j.storageContainer, svc, svc.primary.containerName())
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	if svc.pg == nil && svc.runner == nil && svc.registry == nil {
		return nil
	}

	service, err := j.storageContainer.Services().GetByName(ctx, svc.name)
	if err != nil {
		return rerrors.Wrap(err, "error getting service")
	}

	if svc.pg != nil {
		err = j.backfillPgInstance(ctx, service.ID, svc.name, *svc.pg)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	if svc.runner != nil {
		err = j.backfillRunner(ctx, service.ID, svc.name, *svc.runner)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	if svc.registry != nil {
		err = j.backfillRegistryInstance(ctx, service.ID, svc.name, *svc.registry)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	return nil
}

// registerRunningDeployment records the service and a RUNNING deployment built
// from its live primary container. specName is the specification's unique key:
// it must not collide with a specification already stored for the service.
func registerRunningDeployment(
	ctx context.Context, storageContainer *storage.Container, svc labeledService, specName string,
) error {
	err := storageContainer.Services().UpsertService(ctx, svc.name, svc.displayName)
	if err != nil {
		return rerrors.Wrap(err, "error upserting service")
	}

	service, err := storageContainer.Services().GetByName(ctx, svc.name)
	if err != nil {
		return rerrors.Wrap(err, "error getting service")
	}

	spec := inspectedContainerSpec(
		svc.primary.containerName(), svc.primary.environment, map[string]string{}, svc.primary.info)

	specPayload, err := json.Marshal(spec)
	if err != nil {
		return rerrors.Wrap(err, "error marshaling deployment spec payload")
	}

	specParams := deployments_queries.CreateSpecificationParams{
		Name:        specName,
		ServiceID:   sql.NullInt64{Int64: service.ID, Valid: true},
		VervPayload: pqtype.NullRawMessage{RawMessage: specPayload, Valid: true},
	}

	err = storageContainer.TxManager().Execute(func(tx *sql.Tx) error {
		deployments := storageContainer.Deployments().WithTx(tx)

		specId, txErr := deployments.CreateSpecification(ctx, specParams)
		if txErr != nil {
			return rerrors.Wrap(txErr, "error creating deployment specification")
		}

		deploymentParams := deployments_queries.CreateDeploymentParams{
			NodeID: domain.SelfNodeId,
			Status: deployments_queries.VelezDeploymentStatusRUNNING,
			SpecID: specId,
		}

		_, txErr = deployments.CreateDeployment(ctx, deploymentParams)
		if txErr != nil {
			return rerrors.Wrap(txErr, "error creating deployment")
		}

		return nil
	})
	if err != nil {
		return rerrors.Wrap(err, "error registering service deployment")
	}

	return nil
}

func (j *backfillLabeledServicesJob) backfillPgInstance(
	ctx context.Context, serviceId int64, name string, c labeledContainer,
) error {
	_, err := j.storageContainer.PgInstances().GetPgInstanceByServiceID(ctx, serviceId)
	if err == nil {
		return nil
	}

	if !rerrors.Is(err, user_errors.ErrStorageNotFound) {
		return rerrors.Wrap(err, "error getting pg instance")
	}

	instance := container_derived.PgInstance(name, time.Time{}, c.info.Config.Env)

	password := container_derived.EnvValue(c.info.Config.Env, container_derived.PgaasEnvPassword)

	err = j.ensureSecret(ctx, container_derived.PgInstanceSecretRef(name), password, name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	upsertReq := domain.UpsertPgInstanceReq{
		ServiceId: serviceId,
		DbName:    instance.DbName,
		Username:  instance.Username,
		SecretRef: instance.SecretRef,
		Port:      instance.Port,
	}

	_, err = j.storageContainer.PgInstances().UpsertPgInstance(ctx, upsertReq)
	if err != nil {
		return rerrors.Wrap(err, "error upserting pg instance")
	}

	return nil
}

func (j *backfillLabeledServicesJob) backfillRunner(
	ctx context.Context, serviceId int64, name string, c labeledContainer,
) error {
	_, err := j.storageContainer.Runners().GetRunnerByServiceID(ctx, serviceId)
	if err == nil {
		return nil
	}

	if !rerrors.Is(err, user_errors.ErrStorageNotFound) {
		return rerrors.Wrap(err, "error getting runner")
	}

	runner := container_derived.Runner(name, time.Time{}, c.info.Config.Labels)

	if container_derived.IsGitlabRunner(runner) {
		runner.Concurrent = j.readGitlabConcurrent(ctx, c.containerId(), name)
	}

	registrationToken := container_derived.EnvValue(c.info.Config.Env, container_derived.RunnerRegistrationTokenEnvVar)

	err = j.ensureSecret(ctx, domain.RunnerRegistrationTokenSecretRef(name), registrationToken, name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	// The access token lives nowhere on the container.
	err = j.ensureSecret(ctx, domain.RunnerAccessTokenSecretRef(name), "", name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	upsertReq := domain.UpsertRunnerReq{
		ServiceID:  serviceId,
		Provider:   runner.Provider,
		Scope:      runner.Scope,
		Target:     runner.Target,
		Labels:     runner.Labels,
		SecretRef:  runner.SecretRef,
		BaseUrl:    runner.BaseUrl,
		Concurrent: runner.Concurrent,
	}

	_, err = j.storageContainer.Runners().UpsertRunner(ctx, upsertReq)
	if err != nil {
		return rerrors.Wrap(err, "error upserting runner")
	}

	return nil
}

func (j *backfillLabeledServicesJob) backfillRegistryInstance(
	ctx context.Context, serviceId int64, name string, c labeledContainer,
) error {
	_, err := j.storageContainer.RegistryInstances().GetRegistryInstanceByServiceID(ctx, serviceId)
	if err == nil {
		return nil
	}

	if !rerrors.Is(err, user_errors.ErrStorageNotFound) {
		return rerrors.Wrap(err, "error getting registry instance")
	}

	instance := container_derived.RegistryInstance(name, time.Time{}, c.info.Config.Labels)

	// The password lives in an htpasswd file inside the container.
	err = j.ensureSecret(ctx, container_derived.RegistryInstanceSecretRef(name), "", name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	upsertReq := domain.UpsertRegistryInstanceReq{
		ServiceId: serviceId,
		Port:      instance.Port,
		UiPort:    instance.UiPort,
		Username:  instance.Username,
		SecretRef: instance.SecretRef,
	}

	_, err = j.storageContainer.RegistryInstances().UpsertRegistryInstance(ctx, upsertReq)
	if err != nil {
		return rerrors.Wrap(err, "error upserting registry instance")
	}

	return nil
}

// ensureSecret stores value under ref unless the ref already holds one. An
// empty value with nothing stored is not an error: the row is still written
// and the operator is warned the secret must be supplied again.
func (j *backfillLabeledServicesJob) ensureSecret(
	ctx context.Context, ref domain.SecretRef, value, serviceName string,
) error {
	_, err := j.secrets.Get(ctx, ref)
	if err == nil {
		return nil
	}

	if !rerrors.Is(err, user_errors.ErrSecretNotFound) {
		return rerrors.Wrap(err, "error getting secret")
	}

	if value == "" {
		log.Ctx(ctx).Warn().
			Str("service", serviceName).
			Str("secret_ref", ref.String()).
			Msg("secret cannot be recovered from the container, it must be supplied again")

		return nil
	}

	err = j.secrets.Put(ctx, ref, value)
	if err != nil {
		return rerrors.Wrap(err, "error putting secret")
	}

	return nil
}

func (j *backfillLabeledServicesJob) readGitlabConcurrent(ctx context.Context, containerId, name string) int32 {
	config, err := dockerutils.ReadFromContainer(ctx, j.dockerAPI, containerId, gitlab_runner_config.ConfigPath)
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
