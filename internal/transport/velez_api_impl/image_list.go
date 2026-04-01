package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	velez_api "go.vervstack.ru/Velez/internal/api/server/api/grpc"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) SearchImages(ctx context.Context, req *velez_api.SearchImages_Request) (
	*velez_api.SearchImages_Response, error) {

	searchReq := domain.ImageSearchRequest{
		Term:            req.Name,
		UseOfficialOnly: toolbox.FromPtr(req.UseOnlyOfficial),
	}

	_, err := dockerutils.SearchImages(ctx, impl.dockerAPI, searchReq)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return nil, nil
}
