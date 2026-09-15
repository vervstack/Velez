package registryaas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func (s *RegistryaasService) GetRegistryInstanceCredentials(
	ctx context.Context, name string,
) (domain.RegistryInstanceCredentials, error) {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return domain.RegistryInstanceCredentials{}, rerrors.Wrap(err, "error getting registry instance service")
	}

	instance, err := s.dataStorage.RegistryInstances().GetRegistryInstanceByServiceID(ctx, svc.ID)
	if err != nil {
		return domain.RegistryInstanceCredentials{}, rerrors.Wrap(err, "error getting registry instance row")
	}

	secretRef, err := domain.ParseSecretRef(instance.SecretRef)
	if err != nil {
		return domain.RegistryInstanceCredentials{}, rerrors.Wrap(err, "error parsing registry instance secret ref")
	}

	password, err := s.secrets.Get(ctx, secretRef)
	if err != nil {
		return domain.RegistryInstanceCredentials{}, rerrors.Wrap(err, "error resolving registry instance password")
	}

	registryUrl, err := s.registryUrl(ctx, name)
	if err != nil {
		return domain.RegistryInstanceCredentials{}, err
	}

	credentials := domain.RegistryInstanceCredentials{
		Username:    instance.Username,
		Password:    password,
		RegistryUrl: registryUrl,
	}

	return credentials, nil
}

// registryUrl reads the url a registerRegistryRowJob-equivalent step already
// resolved and persisted into velez.registries at creation time, rather than
// recomputing it - the exposed port used to build it isn't itself stored on
// domain.RegistryInstance (only the fixed container-internal port is; see
// that type's doc comment).
func (s *RegistryaasService) registryUrl(ctx context.Context, name string) (string, error) {
	all, err := s.dataStorage.Registries().ListRegistries(ctx)
	if err != nil {
		return "", rerrors.Wrap(err, "error listing registries")
	}

	for _, reg := range all {
		if reg.Name == name {
			return reg.Url, nil
		}
	}

	return "", rerrors.Wrap(user_errors.ErrRegistryNotFound)
}
