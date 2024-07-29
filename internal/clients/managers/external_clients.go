package managers

import (
	"context"

	"github.com/godverv/matreshka-be/pkg/matreshka_api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/clients/configurator"
	grpcClients "github.com/godverv/Velez/internal/clients/grpc"
	"github.com/godverv/Velez/internal/config"
)

type externalClients struct {
	configurator clients.Configurator

	matreshka matreshka_api.MatreshkaBeAPIClient
}

func NewExternalClients(ctx context.Context, cfg config.Config, intCls clients.InternalClients) (clients.ExternalClients, error) {
	var err error
	exCls := &externalClients{}

	// Matreshka
	{
		logrus.Debug("Initializing matreshka client")
		exCls.matreshka, err = grpcClients.NewMatreshkaBeAPIClient(ctx, cfg, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logrus.Fatalf("error getting matreshka api: %s", err)
		}
	}

	// Configurator
	{
		logrus.Debug("Initializing configuration manager")
		exCls.configurator = configurator.New(exCls.matreshka, intCls.Docker())
	}

	return exCls, nil
}

func (c *externalClients) Configurator() clients.Configurator {
	return c.configurator
}
