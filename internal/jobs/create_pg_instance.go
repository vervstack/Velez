package jobs

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/service_manager/pgaas/pgdeploy"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	CreatePgInstanceAction = "create_pg_instance"

	stepDeployPg              = "deploy_pg"
	stepWaitPg                = "wait_pg"
	stepRegisterPgInstanceRow = "register_pg_instance_row"
	stepBindPgOwnerService    = "bind_owner_service"

	pgDeployWaitTimeout = 180 * time.Second
)

type createPgInstanceHandler struct {
	dataStorage  storage.Storage
	secretsStore secrets.Store
	vervServices service.VervServicesService
	jobsEngine   Engine
}

func NewCreatePgInstanceHandler(
	dataStorage storage.Storage,
	secretsStore secrets.Store,
	vervServices service.VervServicesService,
	jobsEngine Engine,
) TaskHandler {
	return &createPgInstanceHandler{
		dataStorage:  dataStorage,
		secretsStore: secretsStore,
		vervServices: vervServices,
		jobsEngine:   jobsEngine,
	}
}

func (h *createPgInstanceHandler) Action() string {
	return CreatePgInstanceAction
}

func (h *createPgInstanceHandler) NewContext() TaskContext {
	return &velez_api.CreatePgInstanceTaskPayload{}
}

func (h *createPgInstanceHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.CreatePgInstanceTaskPayload)
	if !ok {
		panic("create_pg_instance: BuildJobs called with mismatched TaskContext type")
	}

	request := payload.GetRequest()
	instanceName := labels.PgaasNamePrefix + request.GetName()
	secretRef := pgSecretRef(request.GetName())

	namedJobs := []NamedJob{
		{
			Name: stepGenerateCredentials,
			Job: &generatePgCredentialsJob{
				secrets:   h.secretsStore,
				secretRef: secretRef,
			},
		},
		{
			Name: stepDeployPg,
			Job: &deployPgInstanceJob{
				boxes:        h.dataStorage.ResourceBoxes(),
				vervServices: h.vervServices,
				secrets:      h.secretsStore,
				secretRef:    secretRef,
				req:          request,
				instanceName: instanceName,
			},
		},
		{
			Name: stepWaitPg,
			Job: &waitForPgDeployJob{
				jobsEngine:   h.jobsEngine,
				environment:  request.GetEnvironment(),
				instanceName: instanceName,
			},
		},
		{
			Name: stepRegisterPgInstanceRow,
			Job: &registerPgInstanceRowJob{
				services:     h.dataStorage.Services(),
				pgInstances:  h.dataStorage.PgInstances(),
				secretRef:    secretRef,
				dbName:       pgdeploy.SanitizeIdentifier(request.GetName()),
				instanceName: instanceName,
			},
		},
	}

	if request.GetOwnerService() != "" {
		bindJob := &bindPgOwnerServiceJob{
			serviceResources: h.dataStorage.ServiceResources(),
			ownerService:     request.GetOwnerService(),
			instanceName:     instanceName,
		}

		namedJobs = append(namedJobs, NamedJob{Name: stepBindPgOwnerService, Job: bindJob})
	}

	return namedJobs
}

func pgSecretRef(name string) domain.SecretRef {
	return domain.SecretRef{Scope: pgdeploy.SecretScope, Owner: name, Key: pgdeploy.SecretKey}
}

type generatePgCredentialsJob struct {
	secrets   secrets.Store
	secretRef domain.SecretRef
}

func (j *generatePgCredentialsJob) Do(ctx context.Context) error {
	_, err := j.secrets.Get(ctx, j.secretRef)
	if err == nil {
		return nil
	}

	if !rerrors.Is(err, user_errors.ErrSecretNotFound) {
		return rerrors.Wrap(err, "error checking pg instance password")
	}

	password := string(toolbox.RandomBase64(pgdeploy.PasswordLength))

	err = j.secrets.Put(ctx, j.secretRef, password)
	if err != nil {
		return rerrors.Wrap(err, "error storing pg instance password")
	}

	return nil
}

type deployPgInstanceJob struct {
	boxes        storage.ResourceBoxesStorage
	vervServices service.VervServicesService
	secrets      secrets.Store
	secretRef    domain.SecretRef
	req          *velez_api.CreatePgInstance_Request
	instanceName string
}

func (j *deployPgInstanceJob) Do(ctx context.Context) error {
	password, err := j.secrets.Get(ctx, j.secretRef)
	if err != nil {
		return rerrors.Wrap(err, "error reading pg instance password")
	}

	dbName := pgdeploy.SanitizeIdentifier(j.req.GetName())

	creds := pgdeploy.Credentials{
		DbName:   dbName,
		Username: dbName + "_user",
		Password: password,
	}

	deployInput := domain.CreatePgInstanceReq{
		Name:         j.req.GetName(),
		Environment:  j.req.GetEnvironment(),
		Box:          j.req.GetBox(),
		ExposeToPort: j.req.GetExposeToPort(),
		OwnerService: j.req.GetOwnerService(),
	}

	descriptor, smerdRequest, err := pgdeploy.BuildDeployRequest(ctx, j.boxes, deployInput, creds)
	if err != nil {
		return rerrors.Wrap(err, "error building pg instance deploy request")
	}

	deployReq := domain.CreateDeployReq{
		ServiceName:    j.instanceName,
		VervDescriptor: &descriptor,
		LaunchSmerd:    domain.LaunchSmerd{CreateSmerd_Request: smerdRequest},
	}

	err = j.vervServices.CreateNewDeploy(ctx, deployReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating pg instance deploy")
	}

	return nil
}

type waitForPgDeployJob struct {
	jobsEngine   taskWatcher
	environment  string
	instanceName string
}

func (j *waitForPgDeployJob) Do(ctx context.Context) error {
	entityID := SmerdEntityID(j.environment, j.instanceName)

	watchCtx, cancel := context.WithTimeout(ctx, pgDeployWaitTimeout)
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
			"timed out waiting for pg instance container to deploy, last status: %q",
			finalTask.Status,
		)
	}

	if isFailed {
		return rerrors.Wrap(user_errors.ErrTaskFailed, finalTask.Error.String)
	}

	return nil
}

type registerPgInstanceRowJob struct {
	services     storage.ServicesStorage
	pgInstances  storage.PgInstancesStorage
	secretRef    domain.SecretRef
	dbName       string
	instanceName string
}

func (j *registerPgInstanceRowJob) Do(ctx context.Context) error {
	svc, err := j.services.GetByName(ctx, j.instanceName)
	if err != nil {
		return rerrors.Wrap(err, "error getting pg instance service")
	}

	upsertReq := domain.UpsertPgInstanceReq{
		ServiceId: svc.ID,
		DbName:    j.dbName,
		Username:  j.dbName + "_user",
		SecretRef: j.secretRef.String(),
		Port:      pgdeploy.DefaultPort,
	}

	_, err = j.pgInstances.UpsertPgInstance(ctx, upsertReq)
	if err != nil {
		return rerrors.Wrap(err, "error upserting pg instance row")
	}

	return nil
}

type bindPgOwnerServiceJob struct {
	serviceResources storage.ServiceResourcesStorage
	ownerService     string
	instanceName     string
}

func (j *bindPgOwnerServiceJob) Do(ctx context.Context) error {
	err := j.serviceResources.UpsertResource(ctx, j.ownerService, j.instanceName, pgdeploy.ResourceType)
	if err != nil {
		return rerrors.Wrap(err, "error binding pg instance to owner service")
	}

	return nil
}
