package runneraas

import (
	"context"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// Provider mints a short-lived runner registration token from a caller-
// supplied access token and maps registration facts onto a runner
// container's env - the one seam that differs between CI systems. Everything
// else in this package is provider-agnostic; RunnersAPI/RunneraasService
// never branch on provider outside of picking one from this map.
type Provider interface {
	MintRegistrationToken(ctx context.Context, scope velez_api.RunnerScope, target, accessToken string) (string, error)
	RegistrationEnv(
		scope velez_api.RunnerScope, target, runnerName, registrationToken string, labels []string,
	) map[string]string
	DescriptorName() string
	DataPath() string
}
