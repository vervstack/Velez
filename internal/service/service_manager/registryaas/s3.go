package registryaas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
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
