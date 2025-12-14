package vcn_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	api "go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) DeleteNamespace(ctx context.Context, req *api.DeleteVcnNamespace_Request) (*api.DeleteVcnNamespace_Response, error) {
	err := impl.vpnService.DeleteNamespace(ctx, req.Id)
	if err != nil {
		return nil, rerrors.Wrap(err, "")
	}

	return &api.DeleteVcnNamespace_Response{}, nil
}
