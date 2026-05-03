package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (v *VervService) StopService(ctx context.Context, name string) error {
	listReq := &velez_api.ListSmerds_Request{
		Name: &name,
	}

	resp, err := v.containerService.ListSmerds(ctx, listReq)
	if err != nil {
		return rerrors.Wrap(err, "error listing smerds for service")
	}

	for _, smerd := range resp.Smerds {
		err = v.docker.Stop(ctx, smerd.Uuid)
		if err != nil {
			return rerrors.Wrap(err, "error stopping smerd")
		}
	}

	return nil
}

func (v *VervService) RestartService(ctx context.Context, name string) error {
	listReq := &velez_api.ListSmerds_Request{
		Name: &name,
	}

	resp, err := v.containerService.ListSmerds(ctx, listReq)
	if err != nil {
		return rerrors.Wrap(err, "error listing smerds for service")
	}

	for _, smerd := range resp.Smerds {
		err = v.docker.Restart(ctx, smerd.Uuid)
		if err != nil {
			return rerrors.Wrap(err, "error restarting smerd")
		}
	}

	return nil
}
