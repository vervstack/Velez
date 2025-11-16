package vpn_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) ListNamespaces(ctx context.Context, _ *velez_api.ListVpnNamespaces_Request) (
	*velez_api.ListVpnNamespaces_Response, error) {
	namespaces, err := impl.vpnService.ListNamespaces(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}
	_ = namespaces
	return nil, nil
}
