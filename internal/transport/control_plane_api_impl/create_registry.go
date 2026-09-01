package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

// CreateRegistry persists a new container image registry.
func (impl *Impl) CreateRegistry(
	ctx context.Context,
	req *pb.CreateRegistry_Request,
) (*pb.CreateRegistry_Response, error) {
	serviceReq := domain.CreateRegistryReq{
		Name:      req.GetName(),
		Type:      registryTypeFromPb(req.GetType()),
		Url:       req.GetUrl(),
		Username:  req.GetUsername(),
		Secret:    req.GetSecret(),
		IsDefault: req.GetIsDefault(),
	}

	reg, err := impl.vervServices.CreateRegistry(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating registry")
	}

	resp := &pb.CreateRegistry_Response{
		Registry: registryToPb(reg),
	}

	return resp, nil
}
