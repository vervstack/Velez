package vcn_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) CreateNamespace(ctx context.Context, req *velez_api.CreateVcnNamespace_Request) (
	*velez_api.CreateVcnNamespace_Response, error) {
	namespace, err := impl.vpnService.CreateNamespace(ctx, req.Name)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return &velez_api.CreateVcnNamespace_Response{
		Namespace: namespaceToPb(namespace),
	}, nil
}
