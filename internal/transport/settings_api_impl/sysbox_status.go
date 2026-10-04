package settings_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) GetSysboxStatus(
	ctx context.Context,
	_ *pb.GetSysboxStatus_Request,
) (*pb.GetSysboxStatus_Response, error) {
	status, err := impl.settingsService.GetSysboxStatus(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting sysbox status")
	}

	resp := &pb.GetSysboxStatus_Response{
		OsType:              status.OsType,
		KernelVersion:       status.KernelVersion,
		DockerVersion:       status.DockerVersion,
		IsRootless:          status.IsRootless,
		IsSnap:              status.IsSnap,
		IsRuntimeRegistered: status.IsRuntimeRegistered,
		ContainersTotal:     status.ContainersTotal,
		ContainersOnSysbox:  status.ContainersOnSysbox,
	}

	return resp, nil
}
