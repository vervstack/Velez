package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) ListRegistries(
	ctx context.Context,
	_ *pb.ListRegistries_Request,
) (*pb.ListRegistries_Response, error) {
	list, err := impl.vervServices.ListRegistries(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing registries")
	}

	out := make([]*pb.Registry, 0, len(list))
	for _, reg := range list {
		out = append(out, registryToPb(reg))
	}

	resp := &pb.ListRegistries_Response{
		Registries: out,
	}

	return resp, nil
}

// registryToPb maps the domain type onto the wire type. Shared by every
// registry RPC in this package. It never populates a Secret field - pb.Registry
// deliberately has none, so the plaintext secret can't leak into a response.
func registryToPb(reg domain.Registry) *pb.Registry {
	out := &pb.Registry{
		Id:        reg.Id,
		Name:      reg.Name,
		Type:      registryTypeToPb(reg.Type),
		Url:       reg.Url,
		Username:  reg.Username,
		IsDefault: reg.IsDefault,
	}

	if !reg.CreatedAt.IsZero() {
		out.CreatedAt = timestamppb.New(reg.CreatedAt)
	}

	if !reg.UpdatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(reg.UpdatedAt)
	}

	return out
}

func registryTypeToPb(t domain.RegistryType) pb.RegistryType {
	switch t {
	case domain.RegistryTypeDockerHub:
		return pb.RegistryType_REGISTRY_TYPE_DOCKERHUB
	case domain.RegistryTypeGenericV2:
		return pb.RegistryType_REGISTRY_TYPE_GENERIC_V2
	default:
		return pb.RegistryType_REGISTRY_TYPE_UNSPECIFIED
	}
}

func registryTypeFromPb(t pb.RegistryType) domain.RegistryType {
	switch t {
	case pb.RegistryType_REGISTRY_TYPE_DOCKERHUB:
		return domain.RegistryTypeDockerHub
	case pb.RegistryType_REGISTRY_TYPE_GENERIC_V2:
		return domain.RegistryTypeGenericV2
	default:
		return ""
	}
}
