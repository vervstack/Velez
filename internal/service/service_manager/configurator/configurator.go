package configurator

import (
	"context"

	"github.com/docker/docker/client"
	api "go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
)

type Configurator struct {
	api.MatreshkaBeAPIClient
	dockerAPI client.APIClient

	subscriptionStream api.MatreshkaBeAPI_SubscribeOnChangesClient
	updatesChan        chan domain.ConfigurationPatch
}

func New(
	ctx context.Context,
	cluster cluster_clients.ClusterClients,
	docker node_clients.Docker,
) (c *Configurator, err error) {
	// TODO VERV-128
	//stream, err := matreshka.SubscribeOnChanges(ctx)
	//if err != nil {
	//	return nil, rerrors.Wrap(err, "error subscribing to stream")
	//}
	//closer.Add(stream.CloseSend)

	return &Configurator{
		MatreshkaBeAPIClient: cluster.Configurator(),
		dockerAPI:            docker.Client(),
		// TODO VERV-128
		//subscriptionStream:   stream,
		//updatesChan:          handleSubscriptionStream(stream),
	}, nil
}
