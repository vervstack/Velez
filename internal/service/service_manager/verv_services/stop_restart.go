package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (v *VervService) StopService(ctx context.Context, name, environment string) error {
	listReq := &velez_api.ListSmerds_Request{
		Name:        &name,
		Environment: environment,
	}

	resp, err := v.containerService.ListSmerds(ctx, listReq)
	if err != nil {
		return rerrors.Wrap(err, "error listing smerds for service")
	}

	runtime, err := v.runtimes.Runtime(ctx, environment)
	if err != nil {
		return rerrors.Wrap(err, "error resolving environment")
	}

	for _, smerd := range resp.GetSmerds() {
		err = runtime.Stop(ctx, smerd.GetUuid())
		if err != nil {
			return rerrors.Wrap(err, "error stopping smerd")
		}
	}

	return nil
}

func (v *VervService) RestartService(ctx context.Context, name, environment string) error {
	listReq := &velez_api.ListSmerds_Request{
		Name:        &name,
		Environment: environment,
	}

	resp, err := v.containerService.ListSmerds(ctx, listReq)
	if err != nil {
		return rerrors.Wrap(err, "error listing smerds for service")
	}

	runtime, err := v.runtimes.Runtime(ctx, environment)
	if err != nil {
		return rerrors.Wrap(err, "error resolving environment")
	}

	for _, smerd := range resp.GetSmerds() {
		err = runtime.Restart(ctx, smerd.GetUuid())
		if err != nil {
			return rerrors.Wrap(err, "error restarting smerd")
		}
	}

	return nil
}
