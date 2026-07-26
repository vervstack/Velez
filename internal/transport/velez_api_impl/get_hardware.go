package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) GetHardware(
	context.Context,
	*velez_api.GetHardware_Request,
) (*velez_api.GetHardware_Response, error) {
	resp, err := impl.hardwareManager.GetHardware()
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting hardware")
	}

	return resp, nil
}
