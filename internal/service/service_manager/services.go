package service_manager

import (
	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/client"
	"github.com/godverv/matreshka-be/pkg/matreshka_api"

	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1/port_manager"
	"github.com/godverv/Velez/internal/service/service_manager/hardware_manager_v1"
)

type ServiceManager struct {
	containerManager service.ContainerManager
	hardwareManager  service.HardwareManager
}

func New(
	cfg config.Config,
	docker client.CommonAPIClient,
	configClient matreshka_api.MatreshkaBeAPIClient,
	portManager *port_manager.PortManager,
) (service.Services, error) {
	s := &ServiceManager{}

	var err error

	s.containerManager, err = container_manager_v1.NewContainerManager(
		cfg,
		docker,
		configClient,
		portManager,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating container manager")
	}

	s.hardwareManager = hardware_manager_v1.New()

	return s, nil
}

func (s *ServiceManager) GetContainerManagerService() service.ContainerManager {
	return s.containerManager
}

func (s *ServiceManager) GetHardwareManagerService() service.HardwareManager {
	return s.hardwareManager
}
