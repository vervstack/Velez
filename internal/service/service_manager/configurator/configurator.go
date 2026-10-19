package configurator

import (
	api "go.vervstack.ru/Velez/internal/api/clients/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/domain"
)

type ConfigurationServiceImpl struct {
	api.MatreshkaApiClient

	subscriptionStream api.MatreshkaApi_SubscribeOnChangesClient
	updatesChan        chan domain.ConfigurationPatch
}

func New(cluster cluster_clients.ClusterClients) (c *ConfigurationServiceImpl, err error) {
	return &ConfigurationServiceImpl{
		MatreshkaApiClient: cluster.Configurator(),
	}, nil
}
