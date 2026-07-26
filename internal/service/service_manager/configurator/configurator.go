package configurator

import (
	api "go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/domain"
)

type Configurator struct {
	api.MatreshkaBeAPIClient

	subscriptionStream api.MatreshkaBeAPI_SubscribeOnChangesClient
	updatesChan        chan domain.ConfigurationPatch
}

func New(cluster cluster_clients.ClusterClients) (c *Configurator, err error) {
	return &Configurator{
		MatreshkaBeAPIClient: cluster.Configurator(),
	}, nil
}
