package auto_update_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) SetServiceAutoUpdate(
	ctx context.Context,
	req *pb.SetServiceAutoUpdate_Request,
) (*pb.SetServiceAutoUpdate_Response, error) {
	resp, err := impl.autoUpdateService.SetService(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error setting service auto update")
	}

	return resp, nil
}
