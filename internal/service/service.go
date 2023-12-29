package service

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"

	"github.com/godverv/Velez/pkg/velez_api"
)

var ErrNoPortsAvailable = errors.New("no ports available")

type ContainerManager interface {
	LaunchSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (*velez_api.Smerd, error)
	ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error)
	DropSmerds(ctx context.Context, req *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error)
}

type HardwareManager interface {
	GetHardware() (*velez_api.GetHardware_Response, error)
}

type Services interface {
	GetContainerManagerService() ContainerManager
	GetHardwareManagerService() HardwareManager
}
