package managers

import (
	"github.com/godverv/matreshka-be/pkg/matreshka_be_api"
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

	matreshka matreshka_be_api.MatreshkaBeAPIClient
}

func NewExternalClients(cfg config.Config, intCls clients.InternalClients) (clients.ExternalClients, error) {
	var err error
	exCls := &externalClients{}

	// Matreshka
	{

		// TODO REMOVE INSECURE CONNECTION
		logrus.Debug("Initializing matreshka client")
		exCls.matreshka, err = grpcClients.NewMatreshkaBeAPIClient(cfg.DataSources.GrpcMatreshkaBe,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
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
