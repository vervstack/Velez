package vcn_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) ConnectUser(ctx context.Context, req *velez_api.ConnectUser_Request) (
	*velez_api.ConnectUser_Response, error) {

	domainReq := domain.RegisterVcnNodeReq{
		Key:      req.Key,
		Username: req.Username,
	}

	err := impl.vpnService.RegisterNode(ctx, domainReq)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return &velez_api.ConnectUser_Response{}, nil
}
