package vpn_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) ConnectService(ctx context.Context, req *velez_api.ConnectService_Request) (
	*velez_api.ConnectService_Response, error) {

	r := domain.ConnectServiceToVcn{
		ServiceName: req.ServiceName,
	}

	runner := impl.pipeliner.ConnectServiceToVpn(r)

	err := runner.Run(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error connecting service to VPN")
	}

	return &velez_api.ConnectService_Response{}, nil
}
