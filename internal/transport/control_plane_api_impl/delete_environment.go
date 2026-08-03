package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

// DeleteEnvironment removes an environment identified by id or by name.
//
// It is a cascading delete: every Docker container, network and volume tagged
// with the environment's suffix is removed first, and only then is the DB row
// dropped - so a failed cleanup leaves the environment (and its resources)
// intact rather than orphaning containers no environment points at anymore.
func (impl *Impl) DeleteEnvironment(
	ctx context.Context,
	req *pb.DeleteEnvironment_Request,
) (*pb.DeleteEnvironment_Response, error) {
	serviceReq := domain.DeleteEnvironmentReq{
		ID:   req.Id,
		Name: req.Name,
	}

	err := impl.vervServices.DeleteEnvironment(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error deleting environment")
	}

	return &pb.DeleteEnvironment_Response{}, nil
}
