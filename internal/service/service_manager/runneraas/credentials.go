package runneraas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers"
)

// GetRunnerCredentials resolves a runner's stored registration token - the
// only RunnersService operation that resolves a secret_ref to its plaintext
// value.
func (s *RunneraasService) GetRunnerCredentials(ctx context.Context, name string) (domain.RunnerCredentials, error) {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return domain.RunnerCredentials{}, rerrors.Wrap(err, "error getting runner service")
	}

	runner, err := s.dataStorage.Runners().GetRunnerByServiceID(ctx, svc.ID)
	if err != nil {
		return domain.RunnerCredentials{}, rerrors.Wrap(err, "error getting runner row")
	}

	tokenSecretRef := domain.RunnerRegistrationTokenSecretRef(name)

	token, err := s.secrets.Get(ctx, tokenSecretRef)
	if err != nil {
		return domain.RunnerCredentials{}, rerrors.Wrap(err, "error resolving runner registration token")
	}

	providerEnum := velez_api.RunnerProvider(velez_api.RunnerProvider_value[runner.Provider])

	_, err = providers.For(providerEnum)
	if err != nil {
		return domain.RunnerCredentials{}, rerrors.Wrap(err, "error resolving runner provider")
	}

	credentials := domain.RunnerCredentials{
		Token:    token,
		Target:   runner.Target,
		Provider: providerEnum,
	}

	return credentials, nil
}
