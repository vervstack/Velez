package service_manager

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/internal/service/service_manager/configurator"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager"
	"github.com/godverv/Velez/internal/service/service_manager/smerd_launcher"
	"github.com/godverv/Velez/pkg/velez_api"
)

type ServiceManager struct {
	*container_manager.ContainerManager
	*smerd_launcher.SmerdLauncher
	*configurator.Configurator
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

	containerManger := container_manager.NewContainerManager(nodeClients, configService)

	sm := &ServiceManager{
		Configurator: configService,

		ContainerManager: containerManger,
		SmerdLauncher:    smerd_launcher.New(nodeClients, configService),
	}

	// TODO VERV-128
	//go handleConfigurationSubscription(configService, sm)

	return sm, nil
}

func handleConfigurationSubscription(configurationService service.ConfigurationService, manager service.Services) {

	ctx := context.Background()

	for patch := range configurationService.GetUpdates() {
		listReq := &velez_api.ListSmerds_Request{
			Name: toolbox.ToPtr(patch.ServiceName),
		}

		smerds, err := manager.ListSmerds(ctx, listReq)
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
