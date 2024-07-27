package service_manager

import (
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1"
	"github.com/godverv/Velez/internal/service/service_manager/smerd_launcher"
)

type ServiceManager struct {
	*container_manager_v1.ContainerManager

	*smerd_launcher.SmerdLauncher
}

func New(cl clients.Clients) service.Services {
	return &ServiceManager{
		ContainerManager: container_manager_v1.NewContainerManager(cl),

		SmerdLauncher: smerd_launcher.New(cl),
	}
}
