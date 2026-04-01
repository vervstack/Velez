package velez_api_impl

import (
	"context"

	velez_api "go.vervstack.ru/Velez/internal/api/server/api/grpc"
)

func (impl *Impl) GetHardware(context.Context, *velez_api.GetHardware_Request) (*velez_api.GetHardware_Response, error) {
	return impl.hardwareManager.GetHardware()
}
