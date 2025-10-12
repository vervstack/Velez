package velez_api_impl

import (
	"context"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) GetHardware(context.Context, *velez_api.GetHardware_Request) (*velez_api.GetHardware_Response, error) {
	return impl.hardwareManager.GetHardware()
}
