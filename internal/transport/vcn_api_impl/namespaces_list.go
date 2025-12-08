package vcn_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) ListNamespaces(ctx context.Context, _ *velez_api.ListVpnNamespaces_Request) (
	*velez_api.ListVpnNamespaces_Response, error) {
	namespaces, err := impl.vpnService.ListNamespaces(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return namespacesToPb(namespaces), nil
}

func namespacesToPb(namespaces []domain.VpnNamespace) *velez_api.ListVpnNamespaces_Response {
	resp := &velez_api.ListVpnNamespaces_Response{
		Namespaces: make([]*velez_api.Namespace, 0, len(namespaces)),
	}

	for _, ns := range namespaces {
		resp.Namespaces = append(resp.Namespaces, namespaceToPb(ns))
	}

	return resp
}

func namespaceToPb(ns domain.VpnNamespace) *velez_api.Namespace {
	return &velez_api.Namespace{
		Id:   ns.Id,
		Name: ns.Name,
	}
}
