package container_registry_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/transport/common"
)

func (impl *Impl) ListRegistryInstances(
	ctx context.Context,
	req *pb.ListRegistryInstances_Request,
) (*pb.ListRegistryInstances_Response, error) {
	serviceReq := domain.ListRegistryInstancesReq{
		Paging: common.FromPaging(req.GetPaging()),
	}

	list, err := impl.containerRegistryService.ListRegistryInstances(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing registry instances")
	}

	out := make([]*pb.RegistryInstance, 0, len(list.Instances))
	for _, instance := range list.Instances {
		out = append(out, registryInstanceToPb(instance))
	}

	resp := &pb.ListRegistryInstances_Response{
		Instances: out,
		Total:     list.Total,
	}

	return resp, nil
}

// registryInstanceToPb maps the domain view onto the wire type. Shared by
// every registry instance RPC in this package. Never populates a password
// field - pb.RegistryInstance deliberately has none; only
// GetRegistryInstanceCredentials resolves the secret.
func registryInstanceToPb(view domain.RegistryInstanceView) *pb.RegistryInstance {
	out := &pb.RegistryInstance{
		Name:        view.Name,
		Port:        uint32(view.Port),
		UiPort:      uint32(view.UiPort),
		Username:    view.Username,
		Environment: view.Environment,
		Status:      view.Status,
	}

	if view.OwnerService != "" {
		out.OwnerService = &view.OwnerService
	}

	if !view.CreatedAt.IsZero() {
		out.CreatedAt = timestamppb.New(view.CreatedAt)
	}

	if !view.UpdatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(view.UpdatedAt)
	}

	return out
}
