package velez_api_impl

import (
	"context"

	velez_api "go.vervstack.ru/Velez/internal/api/server/api/grpc"
)

func (impl *Impl) Version(_ context.Context, _ *velez_api.Version_Request) (*velez_api.Version_Response, error) {
	return &velez_api.Version_Response{
		Version: impl.version,
	}, nil
}
