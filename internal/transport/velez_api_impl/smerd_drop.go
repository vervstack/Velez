package velez_api_impl

import (
	"context"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) DropSmerd(ctx context.Context, req *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error) {
	return impl.smerdService.DropSmerds(ctx, req)
}
