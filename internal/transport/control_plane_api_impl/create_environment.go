package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

// CreateEnvironment persists a new deployment environment. An omitted suffix
// defaults to the environment's name (applied in the service layer, so every
// caller - not just this RPC - gets the same rule).
func (impl *Impl) CreateEnvironment(
	ctx context.Context,
	req *pb.CreateEnvironment_Request,
) (*pb.CreateEnvironment_Response, error) {
	serviceReq := domain.CreateEnvironmentReq{
		Name:   req.GetName(),
		Suffix: req.GetSuffix(),
	}

	env, err := impl.vervServices.CreateEnvironment(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating environment")
	}

	resp := &pb.CreateEnvironment_Response{
		Environment: environmentToPb(env),
	}

	return resp, nil
}
