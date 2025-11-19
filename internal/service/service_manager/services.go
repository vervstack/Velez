package service_manager

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/backservice/headscale"
	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/service_manager/configurator"
	"go.vervstack.ru/Velez/internal/service/service_manager/container_manager"
	"go.vervstack.ru/Velez/internal/service/service_manager/headscale_manager"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type ServiceManager struct {
	containerManager *container_manager.ContainerManager
	configurator     *configurator.Configurator
	vpnService       *headscale_manager.Service

	docker clients.Docker
}

func New(
	ctx context.Context,
	nodeClients clients.NodeClients,
	clusterClients clients.ClusterClients,
) (service.Services, error) {
	configService, err := configurator.New(
		ctx,
		clusterClients.Configurator(),
		nodeClients.Docker(),
	)
	if err != nil {
		return nil, rerrors.Wrap(err, "error initializing configurator")
	}

	headscaleManager, err := headscale_manager.New(ctx, nodeClients.Docker(), headscale.Name)
	if err != nil {
		return nil, rerrors.Wrap(err, "error initializing headscale")
	}

	sm := &ServiceManager{
		configurator:     configService,
		vpnService:       headscaleManager,
		containerManager: container_manager.New(nodeClients, configService),

		docker: nodeClients.Docker(),
	}

	// TODO VERV-128
	//go handleConfigurationSubscription(configService, sm)

	return sm, nil
}

func (s *ServiceManager) SmerdManager() service.ContainerService {
	return s.containerManager
}

func (s *ServiceManager) ConfigurationService() service.ConfigurationService {
	return s.configurator
}

func (s *ServiceManager) Docker() clients.Docker {
	return s.docker
}

func (s *ServiceManager) VervPrivateNetworkService() service.VervPrivateNetworkService {
	return s.vpnService
}

func handleConfigurationSubscription(configurationService service.ConfigurationService, manager service.Services) {

	ctx := context.Background()

	for patch := range configurationService.GetUpdates() {
		listReq := &velez_api.ListSmerds_Request{
			Name: toolbox.ToPtr(patch.ConfigName),
		}

		smerds, err := manager.SmerdManager().ListSmerds(ctx, listReq)
		if err != nil {
			logrus.Error(rerrors.Wrap(err, "error listing smerds for configuration update hook"))
			continue
		}

		if len(smerds.Smerds) != 1 {
			logrus.Error(rerrors.New("unexpected number of smerds for configuration update hook. Expected 1 got %d", len(smerds.Smerds)))
			continue
		}

		// TODO VERV-128

	}
}
