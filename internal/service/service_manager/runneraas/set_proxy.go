package runneraas

import (
	"context"
	"slices"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

// SetRunnerProxy applies the proxy through the runner redeploy overlay, so the
// recreated container keeps its DOCKER_HOST and its DinD network. The DinD's
// hostname is added to the bypass hosts when a proxy is set: the daemon is
// reached by name and must not go through the proxy. A service that is not a
// runner is left untouched and reported as such.
func (s *RunneraasService) SetRunnerProxy(ctx context.Context, request domain.SetServiceProxyReq) (bool, error) {
	svc, err := s.dataStorage.Services().GetByName(ctx, request.ServiceName)
	if err != nil {
		return false, rerrors.Wrap(err, "error getting service")
	}

	runner, err := s.dataStorage.Runners().GetRunnerByServiceID(ctx, svc.ID)
	if rerrors.Is(err, user_errors.ErrStorageNotFound) {
		return false, nil
	}

	if err != nil {
		return false, rerrors.Wrap(err, "error getting runner row")
	}

	overlay, err := s.runnerOverlay(ctx, svc.Env, request.ServiceName, runner)
	if err != nil {
		return false, rerrors.Wrap(err, "error building runner overlay")
	}

	request.ProxyBypassHosts = slices.Clone(request.ProxyBypassHosts)
	if overlay.dindName != "" && request.ProxyUrl != "" {
		request.ProxyBypassHosts = append(request.ProxyBypassHosts, overlay.dindName)
	}

	request.EnvOverrides = overlay.env
	request.ExtraNetworks = overlay.networks
	request.RemoveNetworks = overlay.removedNetworks
	request.LabelOverrides = overlay.labels

	err = s.vervServices.SetServiceProxy(ctx, request)
	if err != nil {
		return false, rerrors.Wrap(err, "error setting runner proxy")
	}

	return true, nil
}
