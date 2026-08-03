package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
)

// resolveEnvironment validates the `environment` field carried by
// CreateService/CreateDeploy requests and maps it onto the Docker suffix.
//
// proto3 has no `required` keyword, so this - plus its velez_api_impl twin -
// is where "environment is required" is actually enforced. Only the six
// request messages that gained the field call it; no other RPC does.
func (impl *Impl) resolveEnvironment(ctx context.Context, environment string) (string, error) {
	suffix, err := impl.servicesService.ResolveEnvironmentSuffix(ctx, environment)
	if err != nil {
		return "", rerrors.Wrap(err, "error resolving environment")
	}

	return suffix, nil
}
