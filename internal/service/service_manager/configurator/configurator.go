package configurator

import (
	"context"

	"github.com/docker/docker/client"
	api "go.vervstack.ru/matreshka/pkg/matreshka_be_api"

	"go.vervstack.ru/Velez/internal/domain"
)

type Configurator struct {
	api.MatreshkaBeAPIClient
	dockerAPI client.CommonAPIClient

	subscriptionStream api.MatreshkaBeAPI_SubscribeOnChangesClient
	updatesChan        chan domain.ConfigurationPatch
}

func New(
	ctx context.Context,
	matreshka api.MatreshkaBeAPIClient,
	docker client.CommonAPIClient,
) (c *Configurator, err error) {
	// TODO VERV-128
	//stream, err := matreshka.SubscribeOnChanges(ctx)
	//if err != nil {
	//	return nil, rerrors.Wrap(err, "error subscribing to stream")
	//}
	//closer.Add(stream.CloseSend)

	return &Configurator{
		MatreshkaBeAPIClient: matreshka,
		dockerAPI:            docker,
		// TODO VERV-128
		//subscriptionStream:   stream,
		//updatesChan:          handleSubscriptionStream(stream),
	}, nil
}
