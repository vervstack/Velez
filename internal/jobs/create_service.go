package jobs

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	CreateServiceAction = "create_service"

	minServiceNameLen = 4

	// allowedServiceNameSymbols is the exact character set the deleted
	// internal/pipelines/steps/service_steps.ValidateServiceName accepted
	// (its literal listed 0-9 twice; as a set that is the same thing).
	allowedServiceNameSymbols = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_abcdefghijklmnopqrstuvwxyz"
)

var (
	ErrInvalidServiceName  = rerrors.New("service name contains invalid characters", codes.InvalidArgument)
	ErrTooShortServiceName = rerrors.New("service name is too short", codes.InvalidArgument)
)

// Accessor interface the create_service jobs need from their TaskContext.
// *velez_api.CreateServiceTaskPayload satisfies it.

type serviceNameAccessor interface {
	GetName() string
}

// ServicesStorageResolver yields the services storage of whatever backend is
// currently live, re-resolved per call so an enable_statefull swap
// (cluster_clients.ClusterStateManagerContainer.Set) is observed without a
// restart. Mirrors verv_services.VervService.environments().
type ServicesStorageResolver interface {
	Services() storage.ServicesStorage
}

type createServiceHandler struct {
	dataStorage ServicesStorageResolver
}

func NewCreateServiceHandler(sr ServicesStorageResolver) TaskHandler {
	return &createServiceHandler{
		dataStorage: sr,
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
			Name: "validate_name",
			Job:  &validateServiceNameJob{name: payload.GetName()},
		},
		{
			Name: "upsert_service",
			Job: &upsertServiceJob{
				dataStorage: h.dataStorage,
				req:         payload,
			},
		},
	}
}

// validateServiceNameJob carries over the validation rules of the deleted
// internal/pipelines/steps/service_steps.ValidateServiceName verbatim - the
// step was the create_service pipeline's only surviving piece, and this was
// its one caller.
type validateServiceNameJob struct {
	name string
}

func (j *validateServiceNameJob) Do(_ context.Context) error {
	if len(j.name) < minServiceNameLen {
		return rerrors.Wrap(ErrTooShortServiceName)
	}

	invalidCharsMap := map[rune]struct{}{}

	var invalidChars []rune

	for _, r := range j.name {
		if strings.ContainsRune(allowedServiceNameSymbols, r) {
			continue
		}

		_, alreadyHave := invalidCharsMap[r]
		if alreadyHave {
			continue
		}

		invalidCharsMap[r] = struct{}{}
		invalidChars = append(invalidChars, r)
	}

	if len(invalidChars) == 0 {
		return nil
	}

	return rerrors.Wrap(ErrInvalidServiceName, string(invalidChars))
}

type upsertServiceJob struct {
	dataStorage ServicesStorageResolver

	req serviceNameAccessor
}

func (j *upsertServiceJob) Do(ctx context.Context) error {
	err := j.dataStorage.Services().UpsertService(ctx, j.req.GetName())
	if err != nil {
		return rerrors.Wrap(err, "error upserting service")
	}

	return nil
}
