package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

// DeleteRegistry removes a registry identified by id or by name. Unlike
// DeleteEnvironment, this is not a cascading delete - a registry owns no
// Docker resources.
func (impl *Impl) DeleteRegistry(
	ctx context.Context,
	req *pb.DeleteRegistry_Request,
) (*pb.DeleteRegistry_Response, error) {
	serviceReq := domain.DeleteRegistryReq{
		ID:   req.Id,
		Name: req.Name,
	}

	err := impl.vervServices.DeleteRegistry(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error deleting registry")
	}

	return &pb.DeleteRegistry_Response{}, nil
}
