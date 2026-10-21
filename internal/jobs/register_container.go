package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"maps"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/sqlc-dev/pqtype"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/domain/register_preview"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	RegisterContainerAction = "register_container"

	stepInspectContainer      = "inspect_container"
	stepWhitelistImage        = "whitelist_image"
	stepUpsertService         = "upsert_service"
	stepBindExistingContainer = "bind_existing_container"
	stepRecreateWithLabels    = "recreate_with_labels"

	registerRecreateTimeout = 300 * time.Second
)

type registerRequestAccessor interface {
	GetContainerId() string
	GetEnvironment() string
	GetServiceName() string
	GetBindMountLinks() []*bindMountLink
	GetPattern() velez_api.ServicePattern
	GetImageTag() string
	GetPorts() []*velez_api.Port
	GetKeepPortMapping() bool
	GetIsElevatedAccessAllowed() bool
}

type registerRecreateAccessor interface {
	registerRequestAccessor
	registerUpgradeAccessor
	registerPgLoginAccessor
}

type registerPendingSecretsAccessor interface {
	GetPgPendingSecretOwner() string
	GetRunnerPendingSecretOwner() string
	GetRegistryPendingSecretOwner() string
}

type registeredContainerAccessor interface {
	GetContainerName() string
	GetImageName() string
	GetSwitchedServiceContainerId() string
}

type registeredContainerSetter interface {
	SetContainerName(v string)
	SetImageName(v string)
	SetRootContainerName(v string)
	SetPreviousRootContainerName(v string)
}

type registeredServiceAccessor interface {
	GetServiceId() int64
	SetServiceId(v int64)
}

type registerContainerHandler struct {
	dataStorage storage.Storage
	jobsEngine  Engine
	runtimes    container_runtime.RuntimeResolver
	secrets     secrets.Store
	addressBook service.AddressBook
}

func NewRegisterContainerHandler(
	dataStorage storage.Storage,
	jobsEngine Engine,
	runtimes container_runtime.RuntimeResolver,
	secretsStore secrets.Store,
	addressBook service.AddressBook,
) TaskHandler {
	return &registerContainerHandler{
		dataStorage: dataStorage,
		jobsEngine:  jobsEngine,
		runtimes:    runtimes,
		secrets:     secretsStore,
		addressBook: addressBook,
	}
}

func (h *registerContainerHandler) Action() string {
	return RegisterContainerAction
}

func (h *registerContainerHandler) NewContext() TaskContext {
	return &velez_api.RegisterContainerTaskPayload{}
}

func (h *registerContainerHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.RegisterContainerTaskPayload)
	if !ok {
		panic("register_container: BuildJobs called with mismatched TaskContext type")
	}

	isPg := payload.GetPattern() == velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES
	isRunner := payload.GetRunnerTarget() != ""
	isRegistry := payload.GetPattern() == velez_api.ServicePattern_SERVICE_PATTERN_REGISTRY

	var runner registerRunnerAccessor

	if isRunner {
		runner = payload
	}

	var registry registerRegistryAccessor

	if isRegistry {
		registry = payload
	}

	jobs := []NamedJob{
		{
			Name: stepInspectContainer,
			Job: &inspectContainerToRegisterJob{
				dataStorage: h.dataStorage,
				runtimes:    h.runtimes,
				secrets:     h.secrets,
				req:         payload,
				root:        payload,
				pending:     payload,
				ctx:         payload,
				group:       payload,
			},
		},
	}

	jobs = h.appendWhitelistJob(jobs, payload)

	if isPg {
		verifyJob := NamedJob{
			Name: stepVerifyPgLogin,
			Job: &verifyPgLoginJob{
				runtimes: h.runtimes,
				secrets:  h.secrets,
				req:      payload,
				pg:       payload,
				ctx:      payload,
			},
		}

		secretJob := NamedJob{
			Name: stepStorePgSecret,
			Job: &storePgSecretJob{
				runtimes: h.runtimes,
				secrets:  h.secrets,
				req:      payload,
				pg:       payload,
			},
		}

		jobs = append(jobs, verifyJob, secretJob)
	}

	if isRunner {
		secretsJob := NamedJob{
			Name: stepStoreRunnerSecrets,
			Job: &storeRunnerSecretsJob{
				runtimes: h.runtimes,
				secrets:  h.secrets,
				req:      payload,
				runner:   payload,
			},
		}

		jobs = append(jobs, secretsJob)
	}

	if isRegistry {
		verifyJob := NamedJob{
			Name: stepVerifyRegistryLogin,
			Job: &verifyRegistryLoginJob{
				runtimes: h.runtimes,
				secrets:  h.secrets,
				req:      payload,
				registry: payload,
				ctx:      payload,
			},
		}

		secretJob := NamedJob{
			Name: stepStoreRegistrySecret,
			Job: &storeRegistrySecretJob{
				secrets:  h.secrets,
				req:      payload,
				registry: payload,
			},
		}

		jobs = append(jobs, verifyJob, secretJob)
	}

	linkJob := NamedJob{
		Name: stepLinkBindMounts,
		Job: &unlessRootRegistered{
			group: payload,
			inner: &linkBindMountsJob{
				runtimes: h.runtimes,
				req:      payload,
			},
		},
	}

	jobs = append(jobs, linkJob)

	upsertJob := NamedJob{
		Name: stepUpsertService,
		Job: &unlessRootRegistered{
			group: payload,
			inner: &upsertRegisteredServiceJob{
				dataStorage: h.dataStorage,
				req:         payload,
				ctx:         payload,
			},
		},
	}

	jobs = append(jobs, upsertJob)

	syncJob := NamedJob{
		Name: stepSyncAddresses,
		Job:  &syncAddressesJob{addressBook: h.addressBook, root: staticRootService(payload.GetServiceName())},
	}

	switchJob := NamedJob{
		Name: stepSwitchRoot,
		Job:  newSwitchRootJob(h.runtimes, payload),
	}

	isRootSwitched := payload.GetPreviousRootContainerId() != ""

	if h.dataStorage.IsStatefull() {
		if isRootSwitched {
			jobs = append(jobs, switchJob)
		}

		bindJob := NamedJob{
			Name: stepBindExistingContainer,
			Job: &unlessRootRegistered{
				group: payload,
				inner: &bindExistingContainerJob{
					dataStorage: h.dataStorage,
					runtimes:    h.runtimes,
					req:         payload,
					container:   payload,
					service:     payload,
					runner:      runner,
					registry:    registry,
				},
			},
		}

		bindSidecarsJob := NamedJob{
			Name: stepBindSidecars,
			Job: &bindSidecarsJob{
				dataStorage: h.dataStorage,
				req:         payload,
				service:     payload,
				sidecars:    payload,
			},
		}

		jobs = append(jobs, bindJob, bindSidecarsJob)

		if isRunner {
			runnerRowJob := NamedJob{
				Name: stepUpsertRunnerRow,
				Job: &upsertRunnerRowJob{
					dataStorage: h.dataStorage,
					req:         payload,
					runner:      payload,
					service:     payload,
				},
			}

			jobs = append(jobs, runnerRowJob)
		}

		if isRegistry {
			registryRowJob := NamedJob{
				Name: stepUpsertRegistryRow,
				Job: &upsertRegistryRowJob{
					dataStorage: h.dataStorage,
					runtimes:    h.runtimes,
					req:         payload,
					registry:    payload,
					service:     payload,
				},
			}

			jobs = append(jobs, registryRowJob)
		}

		if isPg {
			pgInstanceJob := NamedJob{
				Name: stepUpsertPgInstance,
				Job: &upsertPgInstanceJob{
					dataStorage: h.dataStorage,
					runtimes:    h.runtimes,
					secrets:     h.secrets,
					req:         payload,
					pg:          payload,
					service:     payload,
				},
			}

			jobs = append(jobs, pgInstanceJob)
		}

		return append(jobs, syncJob)
	}

	if isRootSwitched {
		return append(jobs, switchJob, syncJob)
	}

	recreateJob := NamedJob{
		Name: stepRecreateWithLabels,
		Job: &unlessRootRegistered{
			group: payload,
			inner: &recreateWithLabelsJob{
				jobsEngine: h.jobsEngine,
				runtimes:   h.runtimes,
				secrets:    h.secrets,
				req:        payload,
				root:       payload,
				container:  payload,
				runner:     runner,
				registry:   registry,
				isCluster:  h.dataStorage.IsStatefull(),
			},
		},
	}

	sidecarsJob := NamedJob{
		Name: stepRecreateSidecars,
		Job:  newRecreateSidecarsJob(h.runtimes, registerSidecarGroup{payload: payload}),
	}

	return append(jobs, recreateJob, sidecarsJob, syncJob)
}

// appendWhitelistJob adds the elevated-access check for every register that recreates the container or, in cluster
// mode, binds it; a single-node root switch changes neither.
func (h *registerContainerHandler) appendWhitelistJob(
	jobs []NamedJob, payload *velez_api.RegisterContainerTaskPayload,
) []NamedJob {
	isRootSwitched := payload.GetPreviousRootContainerId() != ""

	if isRootSwitched && !h.dataStorage.IsStatefull() {
		return jobs
	}

	whitelistJob := NamedJob{
		Name: stepWhitelistImage,
		Job: &unlessRootRegistered{
			group: payload,
			inner: &whitelistImageJob{
				dataStorage: h.dataStorage,
				runtimes:    h.runtimes,
				req:         payload,
				root:        payload,
				container:   payload,
			},
		},
	}

	return append(jobs, whitelistJob)
}

// registeredServiceLabels are the labels that link a container to its service.
func registeredServiceLabels(serviceName string) map[string]string {
	return register_preview.RegisteredServiceLabels(serviceName)
}

// registeredContainerSpec derives the CreateSmerd request that describes the
// running container, its labels overlaid with the service labels.
func registeredContainerSpec(
	name, environment, serviceName string, info container.InspectResponse,
) *velez_api.CreateSmerd_Request {
	spec := inspectedContainerSpec(name, environment, registeredServiceLabels(serviceName), info)

	spec.Settings.ElevatedAccess = parser.ToElevatedAccess(info.HostConfig)

	return spec
}

// inspectedContainerSpec derives the CreateSmerd request that describes the
// running container, its labels overlaid with overlayLabels.
func inspectedContainerSpec(
	name, environment string, overlayLabels map[string]string, info container.InspectResponse,
) *velez_api.CreateSmerd_Request {
	merged := make(map[string]string)

	if info.Config != nil {
		maps.Copy(merged, info.Config.Labels)
	}

	maps.Copy(merged, overlayLabels)

	spec := parser.ToCreateRequest(name, info)

	spec.Environment = environment
	spec.Labels = merged

	return spec
}

func containerBindingExists(
	ctx context.Context,
	bindings storage.ContainerBindingsStorage,
	environment, containerName string,
) (bool, error) {
	list, err := bindings.ListByNode(ctx, domain.SelfNodeId, environment)
	if err != nil {
		return false, rerrors.Wrap(err, "error listing container bindings")
	}

	for _, binding := range list {
		if binding.ContainerName == containerName {
			return true, nil
		}
	}

	return false, nil
}

// inspectContainerToRegisterJob is the first job of every register_container
// task, so its Rollback runs on any later failure: it drops the pending
// secrets the request left behind.
type inspectContainerToRegisterJob struct {
	dataStorage storage.Storage
	runtimes    container_runtime.RuntimeResolver
	secrets     secrets.Store

	req     registerRequestAccessor
	root    registerRootAccessor
	pending registerPendingSecretsAccessor
	ctx     registeredContainerSetter
	group   registeredGroupSetter
}

func (j *inspectContainerToRegisterJob) Rollback(ctx context.Context) error {
	pgErr := deletePendingPgSecret(ctx, j.secrets, j.pending.GetPgPendingSecretOwner())
	runnerErr := deletePendingRunnerSecrets(ctx, j.secrets, j.pending.GetRunnerPendingSecretOwner())
	registryErr := deletePendingRegistrySecret(ctx, j.secrets, j.pending.GetRegistryPendingSecretOwner())

	err := rerrors.Join(pgErr, runnerErr, registryErr)
	if err != nil {
		return rerrors.Wrap(err, "error deleting pending register secrets")
	}

	return nil
}

func (j *inspectContainerToRegisterJob) Do(ctx context.Context) error {
	runtime, err := j.runtimes.Runtime(ctx, j.req.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	info, found, err := runtime.InspectAny(ctx, j.req.GetContainerId())
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	if !found || info.Config == nil {
		return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound)
	}

	_, isSidecar := info.Config.Labels[labels.Sidecar]
	if isSidecar {
		return rerrors.Wrap(user_errors.ErrContainerIsSidecar)
	}

	name := strings.TrimPrefix(info.Name, "/")

	groupOwnerId := j.root.GetRootContainerId()
	if j.root.GetPreviousRootContainerId() != "" {
		groupOwnerId = j.root.GetPreviousRootContainerId()
	}

	rootInfo := info

	if groupOwnerId != "" {
		rootInfo, found, err = runtime.InspectAny(ctx, groupOwnerId)
		if err != nil {
			return rerrors.Wrap(err, "error inspecting network root")
		}

		if !found || rootInfo.Config == nil {
			return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound)
		}
	}

	boundServices, err := boundServiceNames(ctx, j.dataStorage.ContainerBindings(), j.req.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error checking container binding")
	}

	plan, err := planRegisterGroup(ctx, runtime, rootInfo, boundServices)
	if err != nil {
		return rerrors.Wrap(err, "error planning network group")
	}

	switch {
	case j.root.GetPreviousRootContainerId() != "":
		err = j.planRootSwitch(ctx, runtime, &plan, rootInfo, name)
		if err != nil {
			return rerrors.Wrap(err)
		}
	case j.root.GetRootContainerId() != "":
		rootName := strings.TrimPrefix(rootInfo.Name, "/")

		plan.SidecarNames, err = registerSidecarNames(plan, rootName, name, j.dataStorage.IsStatefull())
		if err != nil {
			return rerrors.Wrap(err)
		}

		j.ctx.SetRootContainerName(rootName)
	}

	err = applyGroupPlan(ctx, j.dataStorage, plan, j.group)
	if err != nil {
		return rerrors.Wrap(err)
	}

	isMountsLinked := isServiceMountsLinked(
		j.dataStorage.IsStatefull(), j.root.GetRootContainerId(), j.root.GetPreviousRootContainerId(),
	)
	if !plan.IsRootRegistered && isMountsLinked {
		_, err = resolveBindMountLinks(j.req.GetServiceName(), info.Mounts, j.req.GetBindMountLinks())
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	isPortCheckRequired := !j.dataStorage.IsStatefull() && !plan.IsRootRegistered && !j.req.GetKeepPortMapping()
	if isPortCheckRequired {
		err = j.checkRequestedPortsFree(ctx, runtime, rootInfo)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	j.ctx.SetContainerName(name)
	j.ctx.SetImageName(info.Config.Image)

	return nil
}

// isServiceMountsLinked reports whether the service container's bind mounts become named volumes: always in cluster
// mode, in single-node mode only when the container is recreated in place - a role change keeps its binds.
func isServiceMountsLinked(isStatefull bool, rootContainerId, previousRootContainerId string) bool {
	if isStatefull {
		return true
	}

	return rootContainerId == "" && previousRootContainerId == ""
}

// planRootSwitch records the names of the old and the new network root and the members to adopt. oldRoot is the
// current network owner; serviceName is the name of the service container.
func (j *inspectContainerToRegisterJob) planRootSwitch(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	plan *networkGroupPlan,
	oldRoot container.InspectResponse,
	serviceName string,
) error {
	names := switchRoleNames{
		OldRoot: strings.TrimPrefix(oldRoot.Name, "/"),
		NewRoot: serviceName,
		Service: serviceName,
	}

	if j.root.GetRootContainerId() != "" {
		newRoot, found, err := runtime.InspectAny(ctx, j.root.GetRootContainerId())
		if err != nil {
			return rerrors.Wrap(err, "error inspecting the new network root")
		}

		if !found || newRoot.Config == nil {
			return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound)
		}

		names.NewRoot = strings.TrimPrefix(newRoot.Name, "/")

		j.ctx.SetRootContainerName(names.NewRoot)
	}

	var err error

	plan.SidecarNames, err = switchSidecarNames(*plan, names, j.dataStorage.IsStatefull())
	if err != nil {
		return rerrors.Wrap(err)
	}

	j.ctx.SetPreviousRootContainerName(names.OldRoot)

	return nil
}

func (j *inspectContainerToRegisterJob) checkRequestedPortsFree(
	ctx context.Context, runtime container_runtime.ContainerRuntime, info container.InspectResponse,
) error {
	occupied, err := runtime.ListOccupiedPorts(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error listing occupied ports")
	}

	port, isOccupied := findForeignOccupiedPort(j.req.GetPorts(), occupied, currentHostPorts(info))
	if isOccupied {
		return rerrors.Wrap(user_errors.ErrPortOccupied, strconv.FormatUint(uint64(port), 10))
	}

	return nil
}

type upsertRegisteredServiceJob struct {
	dataStorage storage.Storage

	req registerRequestAccessor
	ctx registeredServiceAccessor
}

func (j *upsertRegisteredServiceJob) Do(ctx context.Context) error {
	name := j.req.GetServiceName()

	err := j.dataStorage.Services().UpsertService(ctx, name, name)
	if err != nil {
		return rerrors.Wrap(err, "error upserting service")
	}

	svc, err := j.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error getting service")
	}

	j.ctx.SetServiceId(svc.ID)

	return nil
}

// bindExistingContainerJob leaves the running container untouched: deploy
// watcher reconciles RUNNING deployments by spec name == container name and
// leaves an already-running container alone.
type bindExistingContainerJob struct {
	dataStorage storage.Storage
	runtimes    container_runtime.RuntimeResolver

	req       registerRequestAccessor
	container registeredContainerAccessor
	service   registeredServiceAccessor
	runner    registerRunnerAccessor
	registry  registerRegistryAccessor
}

func (j *bindExistingContainerJob) Do(ctx context.Context) error {
	bindings := j.dataStorage.ContainerBindings()
	if bindings == nil {
		return rerrors.Wrap(user_errors.ErrContainerBindingsUnavailable)
	}

	containerName := j.container.GetContainerName()

	isBound, err := containerBindingExists(ctx, bindings, j.req.GetEnvironment(), containerName)
	if err != nil {
		return rerrors.Wrap(err, "error checking container binding")
	}

	if isBound {
		return nil
	}

	runtime, err := j.runtimes.Runtime(ctx, j.req.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	serviceContainerId := j.req.GetContainerId()
	if j.container.GetSwitchedServiceContainerId() != "" {
		serviceContainerId = j.container.GetSwitchedServiceContainerId()
	}

	info, found, err := runtime.InspectAny(ctx, serviceContainerId)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	if !found {
		return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound)
	}

	plan, err := resolveBindMountLinks(j.req.GetServiceName(), info.Mounts, j.req.GetBindMountLinks())
	if err != nil {
		return rerrors.Wrap(err)
	}

	spec := registeredContainerSpec(containerName, j.req.GetEnvironment(), j.req.GetServiceName(), info)

	if j.req.GetImageTag() != "" {
		spec.ImageName, err = imageWithTag(spec.GetImageName(), j.req.GetImageTag())
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	spec.Settings.Volumes = registerVolumes(info, plan.linked)
	spec.Settings.BindMounts = plan.kept

	maps.Copy(spec.GetLabels(), registeredPatternLabels(j.req.GetPattern()))

	if j.runner != nil {
		maps.Copy(spec.GetLabels(), registeredRunnerLabels(j.runner))
	}

	if j.registry != nil {
		registryLabels := registeredRegistryLabels(j.registry.GetRegistryUsername(), registryInspectedPort(info))

		maps.Copy(spec.GetLabels(), registryLabels)
	}

	specPayload, err := json.Marshal(spec)
	if err != nil {
		return rerrors.Wrap(err, "error marshaling deployment spec payload")
	}

	serviceId := j.service.GetServiceId()

	specParams := deployments_queries.CreateSpecificationParams{
		Name:        containerName,
		ServiceID:   sql.NullInt64{Int64: serviceId, Valid: true},
		VervPayload: pqtype.NullRawMessage{RawMessage: specPayload, Valid: true},
	}

	binding := domain.ContainerBinding{
		ServiceId:     serviceId,
		ServiceName:   j.req.GetServiceName(),
		NodeId:        domain.SelfNodeId,
		Environment:   j.req.GetEnvironment(),
		ContainerName: containerName,
	}

	err = j.dataStorage.TxManager().Execute(func(tx *sql.Tx) error {
		deployments := j.dataStorage.Deployments().WithTx(tx)

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

		txErr = bindings.WithTx(tx).Upsert(ctx, binding)
		if txErr != nil {
			return rerrors.Wrap(txErr, "error upserting container binding")
		}

		return nil
	})
	if err != nil {
		return rerrors.Wrap(err, "error binding container")
	}

	return nil
}

// recreateWithLabelsJob relies on upgrade_smerd's blue-green swap to
// recreate the container with the service labels.
type recreateWithLabelsJob struct {
	jobsEngine Engine
	runtimes   container_runtime.RuntimeResolver
	secrets    secrets.Store

	req       registerRecreateAccessor
	root      registerRootAccessor
	container registeredContainerAccessor
	runner    registerRunnerAccessor
	registry  registerRegistryAccessor

	isCluster bool
}

func (j *recreateWithLabelsJob) Do(ctx context.Context) error {
	if j.root.GetRootContainerId() != "" {
		err := j.recreateRootAsSidecar(ctx)
		if err != nil {
			return rerrors.Wrap(err)
		}

		return nil
	}

	containerName := j.container.GetContainerName()

	runtime, err := j.runtimes.Runtime(ctx, j.req.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	info, found, err := runtime.InspectAny(ctx, j.req.GetContainerId())
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	if !found || info.Config == nil {
		return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound)
	}

	image := j.container.GetImageName()

	if j.req.GetImageTag() != "" {
		image, err = imageWithTag(image, j.req.GetImageTag())
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	upgradeReq := &velez_api.UpgradeSmerd_Request{
		Name:        containerName,
		Image:       image,
		Environment: j.req.GetEnvironment(),
	}

	extraLabels := registeredLabels(j.req.GetServiceName(), j.req.GetPattern())

	extraLabels[labels.OnboardedFromLabel] = info.ID

	stampElevatedAccessAllowance(extraLabels, j.req.GetIsElevatedAccessAllowed(), j.isCluster)

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: upgradeReq,
		ExtraLabels:    extraLabels,

		IsSidecarsSkipped:  true,
		IsOldContainerKept: true,
	}

	if j.req.GetPattern() == velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES {
		containerEnv := parser.ToDockerEnv(info.Config.Env)

		login, loginErr := loadPgLogin(ctx, j.secrets, containerEnv, j.req, j.req, true)
		if loginErr != nil {
			return rerrors.Wrap(loginErr)
		}

		payload.ExtraEnv = pgMissingEnv(containerEnv, login)
	}

	if j.runner != nil {
		err = j.applyRunnerOverlay(ctx, payload, info.Config.Env)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	err = applyRegisterOverrides(payload, info, j.req)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if j.registry != nil {
		err = j.applyRegistryOverlay(ctx, payload, info)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	err = j.upgradeAndWait(ctx, containerName, payload)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

// upgradeAndWait runs an upgrade_smerd task for the container and waits for it.
func (j *recreateWithLabelsJob) upgradeAndWait(
	ctx context.Context, containerName string, payload *velez_api.UpgradeSmerdTaskPayload,
) error {
	// A dedicated entity id: velez.tasks is UNIQUE (entity_id, action), so the
	// container's plain upgrade id would dedupe onto an earlier upgrade task.
	entityId := SmerdEntityID(j.req.GetEnvironment(), containerName) + "/" + RegisterContainerAction

	_, err := j.jobsEngine.Enqueue(ctx, entityId, UpgradeSmerdAction, payload)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing upgrade_smerd task")
	}

	watchCtx, cancel := context.WithTimeout(ctx, registerRecreateTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	for task := range j.jobsEngine.Watch(watchCtx, entityId, UpgradeSmerdAction) {
		finalTask = task
	}

	isDone := finalTask.Status == tasks_queries.VelezTaskStatusDONE
	isFailed := finalTask.Status == tasks_queries.VelezTaskStatusFAILED

	if isFailed {
		return rerrors.Wrap(user_errors.ErrTaskFailed, finalTask.Error.String)
	}

	if !isDone {
		return rerrors.Wrapf(
			watchCtx.Err(),
			"timed out waiting for upgrade_smerd task, last status: %q",
			finalTask.Status,
		)
	}

	return nil
}
