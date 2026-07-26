package jobs

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/pipelines/steps/service_steps"
	"go.vervstack.ru/Velez/internal/storage"
)

const CreateServiceAction = "create_service"

// Accessor interface the create_service jobs need from their TaskContext.
// *velez_api.CreateServiceTaskPayload satisfies it.

type serviceNameAccessor interface {
	GetName() string
}

type createServiceHandler struct {
	servicesStorage storage.ServicesStorage
}

func NewCreateServiceHandler(servicesStorage storage.ServicesStorage) TaskHandler {
	return &createServiceHandler{
		servicesStorage: servicesStorage,
	}
}

func (h *createServiceHandler) Action() string {
	return CreateServiceAction
}

func (h *createServiceHandler) NewContext() TaskContext {
	return &velez_api.CreateServiceTaskPayload{}
}

func (h *createServiceHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.CreateServiceTaskPayload)
	if !ok {
		panic("create_service: BuildJobs called with mismatched TaskContext type")
	}

	return []NamedJob{
		{
			// Reuses the existing validation rules from the in-memory
			// pipeline step - steps.Step and jobs.Job share the same
			// Do(ctx) error shape, so it plugs in directly.
			Name: "validate_name",
			Job:  service_steps.ValidateServiceName(payload.GetName()),
		},
		{
			Name: "upsert_service",
			Job: &upsertServiceJob{
				servicesStorage: h.servicesStorage,
				req:             payload,
			},
		},
	}
}

type upsertServiceJob struct {
	servicesStorage storage.ServicesStorage

	req serviceNameAccessor
}

func (j *upsertServiceJob) Do(ctx context.Context) error {
	err := j.servicesStorage.UpsertService(ctx, j.req.GetName())
	if err != nil {
		return rerrors.Wrap(err, "error upserting service")
	}

	return nil
}
