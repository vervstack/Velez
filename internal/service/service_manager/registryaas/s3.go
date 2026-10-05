package registryaas

import (
	"context"
	"errors"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/garage"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// readS3Binding reads the registry's S3 backend from its own config. The
// second result is false for a volume-backed registry.
func (s *RegistryaasService) readS3Binding(
	ctx context.Context, name, environment string,
) (domain.RegistryS3Binding, bool, error) {
	env, err := s.configResolver.ReadEnv(ctx, name, environment)
	if err != nil {
		return domain.RegistryS3Binding{}, false, rerrors.Wrap(err, "error reading registry config")
	}

	binding, isS3 := domain.RegistryS3BindingFromEnv(env)

	return binding, isS3, nil
}

// deleteS3Access removes the registry's Garage key and everything Velez kept
// about its S3 backend. The bucket stays, so the registry's data survives.
func (s *RegistryaasService) deleteS3Access(
	ctx context.Context, name, environment string, binding domain.RegistryS3Binding,
) error {
	client, err := s.newGarageClient(ctx, environment, binding.InstanceName)
	if err != nil {
		return err
	}

	err = client.DeleteKey(ctx, binding.AccessKeyId)
	if err != nil {
		return rerrors.Wrap(err, "error deleting registry s3 key")
	}

	keyRef := domain.S3KeySecretRef(binding.InstanceName, binding.AccessKeyId)

	err = s.secrets.Delete(ctx, keyRef)
	if err != nil && !errors.Is(err, user_errors.ErrSecretNotFound) {
		return rerrors.Wrap(err, "error deleting registry s3 key secret")
	}

	err = s.dataStorage.ServiceDependencies().DeleteDependenciesOf(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error deleting registry dependencies")
	}

	err = s.configResolver.Delete(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error deleting registry config")
	}

	return nil
}

func (s *RegistryaasService) newGarageClient(
	ctx context.Context, environment, instance string,
) (*garage.Client, error) {
	adminToken, err := s.secrets.Get(ctx, domain.S3AdminTokenSecretRef(instance))
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading s3 admin token")
	}

	containerRuntime, err := s.runtimes.Runtime(ctx, environment)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving container runtime")
	}

	adminUrl, err := garage.AdminUrl(ctx, containerRuntime, domain.S3ServiceName(instance))
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving garage admin url")
	}

	return garage.New(adminUrl, adminToken), nil
}
