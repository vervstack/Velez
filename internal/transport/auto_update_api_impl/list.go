package auto_update_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) ListAutoUpdateServices(
	ctx context.Context,
	_ *pb.ListAutoUpdateServices_Request,
) (*pb.ListAutoUpdateServices_Response, error) {
	resp, err := impl.autoUpdateService.ListServices(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing auto update services")
	}

	return resp, nil
}
