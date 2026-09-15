package container_registry_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
)

// CreateRegistryInstance enqueues the create_registry_instance task and
// returns immediately with the entity_id/action pair TasksApi.WatchTask
// needs to stream its progress - mirrors control_plane_api_impl.EnablePlugin.
// Callers refetch ListRegistryInstances once the watched task reaches DONE;
// the response never carries the built instance.
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
		EnableUi:     req.GetEnableUi(),
	}

	err := impl.containerRegistryService.CreateRegistryInstance(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating registry instance")
	}

	resp := &pb.CreateRegistryInstance_Response{
		EntityId: req.GetName(),
		Action:   jobs.CreateRegistryInstanceAction,
	}

	return resp, nil
}
