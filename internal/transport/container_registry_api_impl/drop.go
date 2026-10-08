package container_registry_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
)

func (impl *Impl) DropRegistryInstance(
	ctx context.Context,
	req *pb.DropRegistryInstance_Request,
) (*pb.DropRegistryInstance_Response, error) {
	err := impl.containerRegistryService.DropRegistryInstance(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error dropping registry instance")
	}

	resp := &pb.DropRegistryInstance_Response{EntityId: req.GetName(), Action: jobs.DropRegistryInstanceAction}

	return resp, nil
}
