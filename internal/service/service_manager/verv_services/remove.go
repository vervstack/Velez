package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (v *VervService) Remove(ctx context.Context, req domain.RemoveServiceReq) error {
	listReq := &velez_api.ListSmerds_Request{
		Name: &req.Name,
	}

	resp, err := v.containerService.ListSmerds(ctx, listReq)
	if err != nil {
		return rerrors.Wrap(err, "error listing smerds for service")
	}

	if len(resp.GetSmerds()) > 0 {
		if !req.DropRunningInstances {
			return rerrors.New("service has running instances, stop/drop them first or set drop_running_instances")
		}

		uuids := make([]string, 0, len(resp.GetSmerds()))
		for _, smerd := range resp.GetSmerds() {
			uuids = append(uuids, smerd.GetUuid())
		}

		dropReq := &velez_api.DropSmerd_Request{
			Uuids: uuids,
		}

		var dropResp *velez_api.DropSmerd_Response

		dropResp, err = v.containerService.DropSmerds(ctx, dropReq)
		if err != nil {
			return rerrors.Wrap(err, "error dropping smerds for service")
		}

		if len(dropResp.GetFailed()) > 0 {
			return rerrors.New("failed to drop some instances of the service", dropResp.GetFailed())
		}
	}

	err = v.servicesStorage.Delete(ctx, req.Name)
	if err != nil {
		return rerrors.Wrap(err, "error deleting service from storage")
	}

	return nil
}
