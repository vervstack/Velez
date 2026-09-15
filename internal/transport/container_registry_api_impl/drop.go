package container_registry_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) DropRegistryInstance(
	ctx context.Context,
	req *pb.DropRegistryInstance_Request,
) (*pb.DropRegistryInstance_Response, error) {
	err := impl.containerRegistryService.DropRegistryInstance(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error dropping registry instance")
	}

	return &pb.DropRegistryInstance_Response{}, nil
}
