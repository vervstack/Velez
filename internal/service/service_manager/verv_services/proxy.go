package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/proxyenv"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// SetServiceProxy schedules an upgrade of the service's running deployment -
// same image, same spec - that overlays the proxy env onto it; an empty proxy
// url deletes the proxy keys instead. Recreating the container is the whole
// mechanism: Docker cannot change the env of a created one.
func (v *VervService) SetServiceProxy(ctx context.Context, request domain.SetServiceProxyReq) error {
	if request.ServiceName == "" {
		return rerrors.Wrap(user_errors.ErrServiceNameRequiredToFind)
	}

	if request.ProxyUrl != "" && !proxyenv.IsValidUrl(request.ProxyUrl) {
		return rerrors.Wrap(user_errors.ErrServiceProxyUrlInvalid)
	}

	upgradeReq := domain.UpgradeDeployReq{
		ServiceName:  request.ServiceName,
		EnvOverrides: proxyenv.Overrides(request.ProxyUrl, request.ProxyBypassHosts),
	}

	err := v.UpgradeDeploy(ctx, upgradeReq)
	if err != nil {
		return rerrors.Wrap(err, "error upgrading deployment with proxy env")
	}

	return nil
}
