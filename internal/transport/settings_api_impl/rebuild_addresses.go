package settings_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) RebuildAddresses(
	ctx context.Context,
	_ *pb.RebuildAddresses_Request,
) (*pb.RebuildAddresses_Response, error) {
	err := impl.addressBook.RebuildAsync(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error rebuilding addresses")
	}

	resp := &pb.RebuildAddresses_Response{}

	return resp, nil
}
