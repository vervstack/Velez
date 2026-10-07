package settings_api_impl

import (
	"context"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) GetAddressesRebuildStatus(
	_ context.Context,
	_ *pb.GetAddressesRebuildStatus_Request,
) (*pb.GetAddressesRebuildStatus_Response, error) {
	status := impl.addressBook.RebuildStatus()

	resp := &pb.GetAddressesRebuildStatus_Response{
		IsRunning:  status.IsRunning,
		TotalSteps: status.TotalSteps,
		DoneSteps:  status.DoneSteps,
	}

	if status.LastError != "" {
		resp.LastError = &status.LastError
	}

	return resp, nil
}
