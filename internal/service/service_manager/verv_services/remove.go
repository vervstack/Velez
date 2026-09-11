package verv_services

import (
	"context"
	"fmt"
	"strings"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func (v *VervService) Remove(ctx context.Context, req domain.RemoveServiceReq) error {
	listReq := &velez_api.ListSmerds_Request{
		Name:        &req.Name,
		Environment: req.Environment,
	}

	resp, err := v.containerService.ListSmerds(ctx, listReq)
	if err != nil {
		return rerrors.Wrap(err, "error listing smerds for service")
	}

	if len(resp.GetSmerds()) > 0 {
		if !req.DropRunningInstances {
			return user_errors.ErrServiceHasRunningInstances
		}

		uuids := make([]string, 0, len(resp.GetSmerds()))
		for _, smerd := range resp.GetSmerds() {
			uuids = append(uuids, smerd.GetUuid())
		}

		dropReq := &velez_api.DropSmerd_Request{
			Uuids:       uuids,
			Environment: req.Environment,
		}

		var dropResp *velez_api.DropSmerd_Response

		dropResp, err = v.containerService.DropSmerds(ctx, dropReq)
		if err != nil {
			return rerrors.Wrap(err, "error dropping smerds for service")
		}

		if len(dropResp.GetFailed()) > 0 {
			return user_errors.New("failed to drop some instances of the service: " + formatFailedDrops(dropResp.GetFailed()))
		}
	}

	err = v.dataStorage.Services().Delete(ctx, req.Name)
	if err != nil {
		return rerrors.Wrap(err, "error deleting service from storage")
	}

	return nil
}

// formatFailedDrops renders each failed drop's uuid and cause into a single
// message string - DropSmerd_Response_Error carries information that used to
// be silently discarded: rerrors.New's variadic args are metadata only (a
// bare string is appended to the message, a codes.Code sets the grpc code),
// so passing a []*DropSmerd_Response_Error there did nothing.
func formatFailedDrops(failed []*velez_api.DropSmerd_Response_Error) string {
	parts := make([]string, 0, len(failed))

	for _, f := range failed {
		parts = append(parts, fmt.Sprintf("%s: %s", f.GetUuid(), f.GetCause()))
	}

	return strings.Join(parts, "; ")
}
