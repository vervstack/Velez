package registryaas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/registries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// DropRegistryInstance removes both the registry instance and its UI sidecar
// service, then the registry_instances/registries rows and the password
// secret. Two services are dropped (not one, unlike pgaas.DropPgInstance) -
// CreateRegistryInstance deploys a registry container and a distinct UI
// sidecar container under registryaasUiServiceName(name).
func (s *RegistryaasService) DropRegistryInstance(ctx context.Context, name string) error {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error getting registry instance service")
	}

	instance, err := s.dataStorage.RegistryInstances().GetRegistryInstanceByServiceID(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error getting registry instance row")
	}

	removeReq := domain.RemoveServiceReq{Name: name, DropRunningInstances: true, Environment: svc.Env}

	err = s.vervServices.Remove(ctx, removeReq)
	if err != nil {
		return rerrors.Wrap(err, "error removing registry instance service")
	}

	removeUiReq := domain.RemoveServiceReq{
		Name:                 registryaasUiServiceName(name),
		DropRunningInstances: true,
		Environment:          svc.Env,
	}

	err = s.vervServices.Remove(ctx, removeUiReq)
	if err != nil {
		return rerrors.Wrap(err, "error removing registry instance ui service")
	}

	err = s.dataStorage.RegistryInstances().DeleteRegistryInstance(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error deleting registry instance row")
	}

	deleter, ok := s.dataStorage.Registries().(registries.BuiltinRegistryDeleter)
	if !ok {
		return rerrors.Wrap(user_errors.ErrRegistriesStorageMissingBuiltinDelete)
	}

	err = deleter.DeleteBuiltinRegistry(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error deleting registry row")
	}

	secretRef, err := domain.ParseSecretRef(instance.SecretRef)
	if err != nil {
		return rerrors.Wrap(err, "error parsing registry instance secret ref")
	}

	err = s.secrets.Delete(ctx, secretRef)
	if err != nil {
		return rerrors.Wrap(err, "error deleting registry instance secret")
	}

	return nil
}
