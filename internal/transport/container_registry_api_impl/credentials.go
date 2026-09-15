package container_registry_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) GetRegistryInstanceCredentials(
	ctx context.Context,
	req *pb.GetRegistryInstanceCredentials_Request,
) (*pb.GetRegistryInstanceCredentials_Response, error) {
	creds, err := impl.containerRegistryService.GetRegistryInstanceCredentials(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting registry instance credentials")
	}

	resp := &pb.GetRegistryInstanceCredentials_Response{
		Username:    creds.Username,
		Password:    creds.Password,
		RegistryUrl: creds.RegistryUrl,
	}

	return resp, nil
}
