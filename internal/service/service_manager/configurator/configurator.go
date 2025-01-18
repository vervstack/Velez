package configurator

import (
	"context"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox/closer"
	api "go.verv.tech/matreshka-be/pkg/matreshka_be_api"

	"github.com/godverv/Velez/internal/domain"
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
	stream, err := matreshka.SubscribeOnChanges(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error subscribing to stream")
	}
	closer.Add(stream.CloseSend)

	return &Configurator{
		MatreshkaBeAPIClient: matreshka,
		dockerAPI:            docker,
		subscriptionStream:   stream,
		updatesChan:          handleSubscriptionStream(stream),
	}, nil
}
