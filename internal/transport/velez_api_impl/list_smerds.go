package velez_api_impl

import (
	"context"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (a *Impl) ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error) {
	return a.smerdService.ListSmerds(ctx, req)
}
