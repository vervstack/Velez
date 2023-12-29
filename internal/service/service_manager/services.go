package service_manager

import (
	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/client"

	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/internal/service/container_manager_v1"
	"github.com/godverv/Velez/internal/service/hardware_manager_v1"
)

type serviceManager struct {
	service.ContainerManager
	service.HardwareManager
}

func New(cfg config.Config, docker client.CommonAPIClient) (_ service.Services, err error) {
	s := &serviceManager{}

	s.ContainerManager, err = container_manager_v1.NewContainerManager(cfg, docker)
	if err != nil {
		return nil, errors.Wrap(err, "error creating container manager")
	}

	s.HardwareManager = hardware_manager_v1.New()

	return s, nil
}

func (s *serviceManager) GetContainerManagerService() service.ContainerManager {
	return s.ContainerManager
}
func (s *serviceManager) GetHardwareManagerService() service.HardwareManager {
	return s.HardwareManager
}
