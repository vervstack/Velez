package vpn_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	api "go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) DeleteNamespace(ctx context.Context, req *api.DeleteVpnNamespace_Request) (*api.DeleteVpnNamespace_Response, error) {
	err := impl.vpnService.DeleteNamespace(ctx, req.Id)
	if err != nil {
		return nil, rerrors.Wrap(err, "")
	}

	return &api.DeleteVpnNamespace_Response{}, nil
}
