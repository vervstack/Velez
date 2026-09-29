package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"maps"
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
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	RegisterContainerAction = "register_container"

	stepInspectContainer      = "inspect_container"
	stepUpsertService         = "upsert_service"
	stepBindExistingContainer = "bind_existing_container"
	stepRecreateWithLabels    = "recreate_with_labels"

	registerRecreateTimeout = 300 * time.Second
)

type registerRequestAccessor interface {
	GetContainerId() string
	GetEnvironment() string
	GetServiceName() string
}

type registeredContainerAccessor interface {
	GetContainerName() string
	GetImageName() string
}

type registeredContainerSetter interface {
	SetContainerName(v string)
	SetImageName(v string)
}

type registeredServiceAccessor interface {
	GetServiceId() int64
	SetServiceId(v int64)
}

type registerContainerHandler struct {
	dataStorage storage.Storage
	jobsEngine  Engine
	runtimes    container_runtime.RuntimeResolver
}

func NewRegisterContainerHandler(
	dataStorage storage.Storage,
	jobsEngine Engine,
	runtimes container_runtime.RuntimeResolver,
) TaskHandler {
	return &registerContainerHandler{
		dataStorage: dataStorage,
		jobsEngine:  jobsEngine,
		runtimes:    runtimes,
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

	jobs := []NamedJob{
		{
			Name: stepInspectContainer,
			Job: &inspectContainerToRegisterJob{
				dataStorage: h.dataStorage,
				runtimes:    h.runtimes,
				req:         payload,
				ctx:         payload,
			},
		},
		{
			Name: stepUpsertService,
			Job: &upsertRegisteredServiceJob{
				dataStorage: h.dataStorage,
				req:         payload,
				ctx:         payload,
			},
		},
	}

	if h.dataStorage.IsStatefull() {
		bindJob := NamedJob{
			Name: stepBindExistingContainer,
			Job: &bindExistingContainerJob{
				dataStorage: h.dataStorage,
				runtimes:    h.runtimes,
				req:         payload,
				container:   payload,
				service:     payload,
			},
		}

		return append(jobs, bindJob)
	}

	recreateJob := NamedJob{
		Name: stepRecreateWithLabels,
		Job: &recreateWithLabelsJob{
			jobsEngine: h.jobsEngine,
			req:        payload,
			container:  payload,
		},
	}

	return append(jobs, recreateJob)
}

// registeredServiceLabels are the labels that link a container to its service.
func registeredServiceLabels(serviceName string) map[string]string {
	return map[string]string{
		labels.VervServiceLabel: serviceName,
		labels.DisplayNameLabel: serviceName,
	}
}

// registeredContainerSpec derives the CreateSmerd request that describes the
// running container, its labels overlaid with the service labels.
func registeredContainerSpec(
	name, environment, serviceName string, info container.InspectResponse,
) *velez_api.CreateSmerd_Request {
	merged := make(map[string]string)

	if info.Config != nil {
		maps.Copy(merged, info.Config.Labels)
	}

	maps.Copy(merged, registeredServiceLabels(serviceName))

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

type inspectContainerToRegisterJob struct {
	dataStorage storage.Storage
	runtimes    container_runtime.RuntimeResolver

	req registerRequestAccessor
	ctx registeredContainerSetter
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

	if info.Config.Labels[labels.VervServiceLabel] != "" {
		return rerrors.Wrap(user_errors.ErrContainerAlreadyLinked)
	}

	_, isSidecar := info.Config.Labels[labels.Sidecar]
	if isSidecar {
		return rerrors.Wrap(user_errors.ErrContainerIsSidecar)
	}

	name := strings.TrimPrefix(info.Name, "/")

	bindings := j.dataStorage.ContainerBindings()
	if bindings != nil {
		isBound, bindErr := containerBindingExists(ctx, bindings, j.req.GetEnvironment(), name)
		if bindErr != nil {
			return rerrors.Wrap(bindErr, "error checking container binding")
		}

		if isBound {
			return rerrors.Wrap(user_errors.ErrContainerAlreadyLinked)
		}
	}

	j.ctx.SetContainerName(name)
	j.ctx.SetImageName(info.Config.Image)

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

	info, found, err := runtime.InspectAny(ctx, j.req.GetContainerId())
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	if !found {
		return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound)
	}

	spec := registeredContainerSpec(containerName, j.req.GetEnvironment(), j.req.GetServiceName(), info)

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

	req       registerRequestAccessor
	container registeredContainerAccessor
}

func (j *recreateWithLabelsJob) Do(ctx context.Context) error {
	containerName := j.container.GetContainerName()

	upgradeReq := &velez_api.UpgradeSmerd_Request{
		Name:        containerName,
		Image:       j.container.GetImageName(),
		Environment: j.req.GetEnvironment(),
	}

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: upgradeReq,
		ExtraLabels:    registeredServiceLabels(j.req.GetServiceName()),
	}

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
