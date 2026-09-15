package registryaas

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	// createRegistryInstanceWatchTimeout bounds how long CreateRegistryInstance
	// blocks waiting for the create_registry_instance task to reach a
	// terminal status - a safety net against the task engine getting stuck,
	// not something normal operation should hit. Mirrors
	// service_api_impl.CreateService's and velez_api_impl.CreateSmerd's own
	// watch timeouts, sized the same since this task chains one loader
	// container plus two ordinary service deploys - the same order of work
	// as a single CreateSmerd.
	createRegistryInstanceWatchTimeout = 60 * time.Second
)

// CreateRegistryInstance enqueues the multi-step create_registry_instance
// task (credentials, port resolution, htpasswd loader, the registry and its
// UI sidecar deploy, both storage rows - internal/jobs/create_registry_instance.go)
// and watches it to a terminal status, so the RPC still returns the
// fully-built instance synchronously - unlike ListPgInstances/CreatePgInstance,
// this can't be done inline: creating an instance means deploying two
// containers through the ordinary (asynchronous) CreateNewDeploy/deploy
// watcher path, not something a single service-layer call can do by itself.
func (s *RegistryaasService) CreateRegistryInstance(
	ctx context.Context, req domain.CreateRegistryInstanceReq,
) (domain.RegistryInstanceView, error) {
	initialContext := &velez_api.CreateRegistryInstanceTaskPayload{
		Request: registryInstanceRequestToPb(req),
	}

	_, err := s.jobsEngine.Enqueue(ctx, req.Name, jobs.CreateRegistryInstanceAction, initialContext)
	if err != nil {
		return domain.RegistryInstanceView{}, rerrors.Wrap(err, "error enqueuing create registry instance task")
	}

	watchCtx, cancel := context.WithTimeout(ctx, createRegistryInstanceWatchTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	for task := range s.jobsEngine.Watch(watchCtx, req.Name, jobs.CreateRegistryInstanceAction) {
		finalTask = task
	}

	isDone := finalTask.Status == tasks_queries.VelezTaskStatusDONE
	isFailed := finalTask.Status == tasks_queries.VelezTaskStatusFAILED

	if !isDone && !isFailed && watchCtx.Err() != nil {
		return domain.RegistryInstanceView{}, rerrors.Wrapf(
			watchCtx.Err(),
			"timed out waiting for create_registry_instance task, last status: %q",
			finalTask.Status,
		)
	}

	if finalTask.Status == tasks_queries.VelezTaskStatusFAILED {
		return domain.RegistryInstanceView{}, rerrors.Wrap(user_errors.ErrTaskFailed, finalTask.Error.String)
	}

	svc, err := s.dataStorage.Services().GetByName(ctx, req.Name)
	if err != nil {
		return domain.RegistryInstanceView{}, rerrors.Wrap(err, "error getting registry instance service")
	}

	instance, err := s.dataStorage.RegistryInstances().GetRegistryInstanceByServiceID(ctx, svc.ID)
	if err != nil {
		return domain.RegistryInstanceView{}, rerrors.Wrap(err, "error getting registry instance row")
	}

	view := domain.RegistryInstanceView{
		Name:         req.Name,
		Port:         instance.Port,
		UiPort:       instance.UiPort,
		Username:     instance.Username,
		Environment:  req.Environment,
		OwnerService: req.OwnerService,
		CreatedAt:    instance.CreatedAt,
		UpdatedAt:    instance.UpdatedAt,
	}

	return view, nil
}

// registryInstanceRequestToPb converts the domain request into the wire
// request the create_registry_instance task persists as its own context -
// internal/jobs/create_registry_instance.go reads request fields straight off
// *velez_api.CreateRegistryInstance_Request, so the task payload carries the
// proto message rather than a second, parallel domain-shaped copy.
func registryInstanceRequestToPb(req domain.CreateRegistryInstanceReq) *velez_api.CreateRegistryInstance_Request {
	pbReq := &velez_api.CreateRegistryInstance_Request{
		Name: req.Name,
	}

	if req.Environment != "" {
		pbReq.Environment = &req.Environment
	}

	if req.Box != "" {
		pbReq.Box = &req.Box
	}

	if req.ExposeToPort != 0 {
		pbReq.ExposeToPort = &req.ExposeToPort
	}

	if req.OwnerService != "" {
		pbReq.OwnerService = &req.OwnerService
	}

	return pbReq
}
