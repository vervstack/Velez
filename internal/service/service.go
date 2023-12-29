package service

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/client"

	"github.com/godverv/Velez/internal/service/container_manager_v1"
	"github.com/godverv/Velez/internal/service/hardware_manager_v1"
	"github.com/godverv/Velez/pkg/velez_api"
)

type ContainerManager interface {
	LaunchSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (*velez_api.Smerd, error)
	ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error)
	DropSmerds(ctx context.Context, req *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error)
}

type HardwareManager interface {
	GetHardware() (*velez_api.GetHardware_Response, error)
}

type Service interface {
	GetContainerManagerService() ContainerManager
	GetHardwareManagerService() HardwareManager
}

type service struct {
	ContainerManager
	HardwareManager
}

func New(docker client.CommonAPIClient) (_ Service, err error) {
	s := &service{}

	s.ContainerManager, err = container_manager_v1.NewContainerManager(docker)
	if err != nil {
		return nil, errors.Wrap(err, "error creating container manager")
	}

	s.HardwareManager = hardware_manager_v1.New()

	return s, nil
}

func (s *service) GetContainerManagerService() ContainerManager {
	return s.ContainerManager
}
func (s *service) GetHardwareManagerService() HardwareManager {
	return s.HardwareManager
}
