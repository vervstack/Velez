package container_registry_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) CreateRegistryInstance(
	ctx context.Context,
	req *pb.CreateRegistryInstance_Request,
) (*pb.CreateRegistryInstance_Response, error) {
	serviceReq := domain.CreateRegistryInstanceReq{
		Name:         req.GetName(),
		Environment:  req.GetEnvironment(),
		Box:          req.GetBox(),
		ExposeToPort: req.GetExposeToPort(),
		OwnerService: req.GetOwnerService(),
	}

	instance, err := impl.containerRegistryService.CreateRegistryInstance(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating registry instance")
	}

	resp := &pb.CreateRegistryInstance_Response{
		Instance: registryInstanceToPb(instance),
	}

	return resp, nil
}
