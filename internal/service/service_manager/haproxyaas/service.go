package haproxyaas

import (
	"context"
	"encoding/json"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
)

type HaproxyServiceImpl struct {
	dataStorage    storage.Storage
	configResolver service.ServiceConfigResolver
	jobsEngine     jobs.Engine
	runtimes       container_runtime.RuntimeResolver
}

func New(
	dataStorage storage.Storage,
	configResolver service.ServiceConfigResolver,
	jobsEngine jobs.Engine,
	runtimes container_runtime.RuntimeResolver,
) *HaproxyServiceImpl {
	return &HaproxyServiceImpl{
		dataStorage:    dataStorage,
		configResolver: configResolver,
		jobsEngine:     jobsEngine,
		runtimes:       runtimes,
	}
}

func (s *HaproxyServiceImpl) CreateInstance(ctx context.Context, req *velez_api.CreateHaproxyInstance_Request) error {
	err := domain.ValidateInstanceName(req.GetName())
	if err != nil {
		return rerrors.Wrap(err)
	}

	if req.GetMasterPgInstance() == "" {
		return rerrors.Wrap(user_errors.ErrHaproxyInvalidMaster)
	}

	if req.Port != nil && (req.GetPort() == 0 || req.GetPort() > domain.HaproxyMaxPort) {
		return rerrors.Wrap(user_errors.ErrHaproxyInvalidPort)
	}

	_, err = s.dataStorage.Services().GetByName(ctx, domain.HaproxyServiceName(req.GetName()))
	if err == nil {
		return rerrors.Wrap(user_errors.ErrHaproxyInstanceExists)
	}

	if !rerrors.Is(err, user_errors.ErrStorageNotFound) {
		return rerrors.Wrap(err, "error checking HAProxy service name")
	}

	payload := &velez_api.CreateHaproxyInstanceTaskPayload{Request: req}

	_, err = s.jobsEngine.EnqueueReplacing(ctx, req.GetName(), jobs.CreateHaproxyInstanceAction, payload)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing HAProxy create task")
	}

	return nil
}

func (s *HaproxyServiceImpl) ValidateTarget(
	ctx context.Context,
	req *velez_api.ValidateHaproxyTarget_Request,
) (string, error) {
	network, err := jobs.ValidateHaproxyTarget(ctx, s.dataStorage, s.runtimes, req)
	if err != nil {
		return network, rerrors.Wrap(err)
	}

	return network, nil
}

func (s *HaproxyServiceImpl) SelectMaster(ctx context.Context, req *velez_api.SelectHaproxyMaster_Request) error {
	if req.GetMasterPgInstance() == "" {
		return rerrors.Wrap(user_errors.ErrHaproxyInvalidMaster)
	}

	ref, err := s.findInstance(ctx, req.GetName())
	if err != nil {
		return rerrors.Wrap(err)
	}

	if ref.summary.State != "running" {
		return rerrors.Wrap(user_errors.ErrHaproxyInstanceStopped)
	}

	port := publishedPort(ref.summary)
	if port == 0 {
		instance, _, resolveErr := s.ResolveMasterSelection(ctx, req.GetName())
		if resolveErr != nil {
			return rerrors.Wrap(resolveErr)
		}

		port = instance.Port
	}

	payload := &velez_api.SelectHaproxyMasterTaskPayload{
		Request:     req,
		Environment: ref.environment,
		ContainerId: ref.summary.ID,
		Port:        port,
	}

	task, err := s.jobsEngine.EnqueueReplacing(ctx, req.GetName(), jobs.SelectHaproxyMasterAction, payload)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing HAProxy master selection task")
	}

	var active velez_api.SelectHaproxyMasterTaskPayload

	err = json.Unmarshal(task.Context.RawMessage, &active)
	if err != nil {
		return rerrors.Wrap(err, "error decoding HAProxy master selection task")
	}

	if active.GetRequest().GetMasterPgInstance() != req.GetMasterPgInstance() ||
		active.GetContainerId() != payload.GetContainerId() {
		return rerrors.Wrap(user_errors.ErrHaproxySelectionBusy)
	}

	return nil
}

func (s *HaproxyServiceImpl) DropInstance(ctx context.Context, name string) error {
	ref, err := s.findInstance(ctx, name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	payload := &velez_api.DropHaproxyInstanceTaskPayload{Name: name, Environment: ref.environment}

	_, err = s.jobsEngine.EnqueueReplacing(ctx, name, jobs.DropHaproxyInstanceAction, payload)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing HAProxy drop task")
	}

	return nil
}
