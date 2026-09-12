package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

// UpdateEnvironment renames an environment and/or changes its Docker suffix
// or dedicated Docker host. All three are optional - an omitted one is left
// as-is.
func (impl *Impl) UpdateEnvironment(
	ctx context.Context,
	req *pb.UpdateEnvironment_Request,
) (*pb.UpdateEnvironment_Response, error) {
	serviceReq := domain.UpdateEnvironmentReq{
		ID:         req.GetId(),
		Name:       req.Name,
		Suffix:     req.Suffix,
		DockerHost: req.DockerHost,
	}

	env, err := impl.vervServices.UpdateEnvironment(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error updating environment")
	}

	resp := &pb.UpdateEnvironment_Response{
		Environment: environmentToPb(env),
	}

	return resp, nil
}
