package grpc

import (
	"context"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Api) GetHardware(context.Context, *velez_api.GetHardware_Request) (*velez_api.GetHardware_Response, error) {
	return a.hardwareManager.GetHardware()
}
