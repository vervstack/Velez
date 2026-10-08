package registryaas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
)

// DropRegistryInstance enqueues drop_registry_instance, which removes the
// registry service, its UI sidecar service, the storage rows and the password
// secret (internal/jobs/drop_registry_instance.go).
func (s *RegistryaasService) DropRegistryInstance(ctx context.Context, name string) error {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error getting registry instance service")
	}

	_, err = s.dataStorage.RegistryInstances().GetRegistryInstanceByServiceID(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error getting registry instance row")
	}

	payload := &velez_api.DropRegistryInstanceTaskPayload{Name: name}

	_, err = s.jobsEngine.EnqueueReplacing(ctx, name, jobs.DropRegistryInstanceAction, payload)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing drop registry instance task")
	}

	return nil
}
