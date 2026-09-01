package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

// UpdateRegistry changes a subset of a registry's fields - an omitted one is
// left as-is. An empty Secret is not sent by the frontend when the user
// leaves the secret field blank, which is what makes "leave secret blank to
// keep it unchanged" work.
func (impl *Impl) UpdateRegistry(
	ctx context.Context,
	req *pb.UpdateRegistry_Request,
) (*pb.UpdateRegistry_Response, error) {
	serviceReq := domain.UpdateRegistryReq{
		ID:        req.GetId(),
		Name:      req.Name,
		Url:       req.Url,
		Username:  req.Username,
		Secret:    req.Secret,
		IsDefault: req.IsDefault,
	}

	if req.Type != nil {
		regType := registryTypeFromPb(req.GetType())

		serviceReq.Type = &regType
	}

	reg, err := impl.vervServices.UpdateRegistry(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error updating registry")
	}

	resp := &pb.UpdateRegistry_Response{
		Registry: registryToPb(reg),
	}

	return resp, nil
}
