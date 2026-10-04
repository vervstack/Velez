package jobs

import (
	"context"
	"strconv"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	CreateDindAction = "create_dind"

	stepPutIsolationSecret = "put_isolation_secret"
	stepEnsureDindNetwork  = "ensure_network"
	stepDeployDind         = "deploy_dind"
	stepWaitForDindDeploy  = "wait_for_dind_deploy"
	stepRegisterDindRow    = "register_dind_row"

	IsolationSecretSysbox     = "sysbox"
	IsolationSecretPrivileged = "privileged"

	envDindTlsCertDir = "DOCKER_TLS_CERTDIR"

	dindDeployWaitTimeout = 180 * time.Second
)

type createDindRequestAccessor interface {
	GetRequest() *velez_api.CreateDind_Request
}

type createDindHandler struct {
	dataStorage  storage.Storage
	secretsStore secrets.Store
	vervServices service.VervServicesService
	jobsEngine   Engine
	runtimes     container_runtime.RuntimeResolver
}

func NewCreateDindHandler(
	dataStorage storage.Storage,
	secretsStore secrets.Store,
	vervServices service.VervServicesService,
	jobsEngine Engine,
	runtimes container_runtime.RuntimeResolver,
) TaskHandler {
	return &createDindHandler{
		dataStorage:  dataStorage,
		secretsStore: secretsStore,
		vervServices: vervServices,
		jobsEngine:   jobsEngine,
		runtimes:     runtimes,
	}
}

func (h *createDindHandler) Action() string {
	return CreateDindAction
}

func (h *createDindHandler) NewContext() TaskContext {
	return &velez_api.CreateDindTaskPayload{}
}

func (h *createDindHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.CreateDindTaskPayload)
	if !ok {
		panic("create_dind: BuildJobs called with mismatched TaskContext type")
	}

	name := payload.GetRequest().GetName()

	return []NamedJob{
		{
			Name: stepPutIsolationSecret,
			Job: &putDindIsolationSecretJob{
				secrets: h.secretsStore,
				req:     payload,
			},
		},
		{
			Name: stepEnsureDindNetwork,
			Job: &ensureDindNetworkJob{
				runtimes: h.runtimes,
				req:      payload,
				payload:  payload,
			},
		},
		{
			Name: stepDeployDind,
			Job: &deployDindJob{
				vervServices: h.vervServices,
				req:          payload,
			},
		},
		{
			Name: stepWaitForDindDeploy,
			Job: &waitForDindDeployJob{
				jobsEngine: h.jobsEngine,
				req:        payload,
			},
		},
		{
			Name: stepRegisterDindRow,
			Job: &registerDindRowJob{
				services:      h.dataStorage.Services(),
				dindInstances: h.dataStorage.DindInstances(),
				name:          name,
				req:           payload,
			},
		},
	}
}

func isDindSysboxEnabled(request *velez_api.CreateDind_Request) bool {
	if request.IsSysboxEnabled == nil {
		return true
	}

	return request.GetIsSysboxEnabled()
}

type putDindIsolationSecretJob struct {
	secrets secrets.Store

	req createDindRequestAccessor
}

func (j *putDindIsolationSecretJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()

	isolation := IsolationSecretPrivileged
	if isDindSysboxEnabled(request) {
		isolation = IsolationSecretSysbox
	}

	err := j.secrets.Put(ctx, domain.ContainerIsolationSecretRef(request.GetName()), isolation)
	if err != nil {
		return rerrors.Wrap(err, "error putting container isolation secret")
	}

	return nil
}

type ensureDindNetworkJob struct {
	runtimes container_runtime.RuntimeResolver

	req     createDindRequestAccessor
	payload *velez_api.CreateDindTaskPayload
}

func (j *ensureDindNetworkJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()
	networkName := domain.DindNetworkName(request.GetName())

	containerRuntime, err := j.runtimes.Runtime(ctx, request.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	err = containerRuntime.CreateNetwork(ctx, networkName)
	if err != nil {
		return rerrors.Wrapf(err, "error creating network: %s", networkName)
	}

	j.payload.NetworkName = networkName
	j.payload.Address = domain.DindAddress(request.GetName())

	return nil
}

type deployDindJob struct {
	vervServices service.VervServicesService

	req createDindRequestAccessor
}

func (j *deployDindJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()
	name := request.GetName()

	smerdRequest := buildDindDeployRequest(request)

	deployReq := domain.CreateDeployReq{
		ServiceName: name,
		DisplayName: name,
		LaunchSmerd: domain.LaunchSmerd{CreateSmerd_Request: smerdRequest},
	}

	err := j.vervServices.CreateNewDeploy(ctx, deployReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating dind deploy")
	}

	return nil
}

func buildDindDeployRequest(request *velez_api.CreateDind_Request) *velez_api.CreateSmerd_Request {
	name := request.GetName()

	volume := &velez_api.Volume{
		VolumeName:    domain.DindDataVolumeName(name),
		ContainerPath: domain.DindDataVolumePath,
	}

	networkBind := &velez_api.NetworkBind{
		NetworkName: domain.DindNetworkName(name),
		Aliases:     []string{name},
	}

	settings := &velez_api.Container_Settings{
		Volumes: []*velez_api.Volume{volume},
		Network: []*velez_api.NetworkBind{networkBind},
	}

	restart := &velez_api.RestartPolicy{Type: velez_api.RestartPolicyType_unless_stopped}

	return &velez_api.CreateSmerd_Request{
		Name:         name,
		ImageName:    domain.DindImage,
		Environment:  request.GetEnvironment(),
		Env:          map[string]string{envDindTlsCertDir: ""},
		Labels:       dindLabels(request),
		Settings:     settings,
		Restart:      restart,
		IgnoreConfig: true,
	}
}

func dindLabels(request *velez_api.CreateDind_Request) map[string]string {
	return map[string]string{
		labels.VervServiceLabel:  request.GetName(),
		labels.DindInstanceLabel: vervConfigLabelEnabled,
		labels.DindSysboxLabel:   strconv.FormatBool(isDindSysboxEnabled(request)),
	}
}

type waitForDindDeployJob struct {
	jobsEngine taskWatcher

	req createDindRequestAccessor
}

func (j *waitForDindDeployJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()
	entityID := SmerdEntityID(request.GetEnvironment(), request.GetName())

	watchCtx, cancel := context.WithTimeout(ctx, dindDeployWaitTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	for task := range j.jobsEngine.Watch(watchCtx, entityID, CreateSmerdAction) {
		finalTask = task
	}

	isDone := finalTask.Status == tasks_queries.VelezTaskStatusDONE
	isFailed := finalTask.Status == tasks_queries.VelezTaskStatusFAILED

	if !isDone && !isFailed && watchCtx.Err() != nil {
		return rerrors.Wrapf(
			watchCtx.Err(),
			"timed out waiting for dind container to deploy, last status: %q",
			finalTask.Status,
		)
	}

	if isFailed {
		return rerrors.Wrap(user_errors.ErrTaskFailed, finalTask.Error.String)
	}

	return nil
}

type registerDindRowJob struct {
	services      storage.ServicesStorage
	dindInstances storage.DindInstancesStorage
	name          string

	req createDindRequestAccessor
}

func (j *registerDindRowJob) Do(ctx context.Context) error {
	svc, err := j.services.GetByName(ctx, j.name)
	if err != nil {
		return rerrors.Wrap(err, "error getting dind service")
	}

	upsertReq := domain.UpsertDindInstanceReq{
		ServiceId:       svc.ID,
		IsSysboxEnabled: isDindSysboxEnabled(j.req.GetRequest()),
	}

	_, err = j.dindInstances.UpsertDindInstance(ctx, upsertReq)
	if err != nil {
		return rerrors.Wrap(err, "error upserting dind instance row")
	}

	return nil
}
