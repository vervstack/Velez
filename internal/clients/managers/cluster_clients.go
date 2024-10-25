package managers

import (
	"github.com/godverv/makosh/pkg/makosh_be"
	"github.com/godverv/matreshka-be/pkg/matreshka_be_api"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/clients/configurator"
	"github.com/godverv/Velez/internal/config"
)

type externalClients struct {
	configurator clients.Configurator

	makosh    makosh_be.MakoshBeAPIClient
	matreshka matreshka_be_api.MatreshkaBeAPIClient
}

func NewClusterClients(cfg config.Config, intCls clients.NodeClients) (clients.ClusterClients, error) {
	var err error
	exCls := &externalClients{}
	// Makosh
	{
		//c.GrpcMakosh, err = makosh.New(makoshBackgroundTask.AuthToken,
		//	grpc2.WithTransportCredentials(insecure.NewCredentials()))
		//if err != nil {
		//	logrus.Fatalf("error creating makosh client %s", errors.Wrap(err))
		//}

	}

	// Matreshka
	{

		// TODO REMOVE INSECURE CONNECTION
		//logrus.Debug("Initializing matreshka client")
		//exCls.matreshka, err = grpcClients.NewMatreshkaBeAPIClient(cfg.Environment.MatreshkaPort,
		//	grpc.WithTransportCredentials(insecure.NewCredentials()))
		//if err != nil {
		//	logrus.Fatalf("error getting matreshka api: %s", err)
		//}
	}

	// Configurator
	{
		logrus.Debug("Initializing configuration manager")
		exCls.configurator = configurator.New(exCls.matreshka, intCls.Docker())
	}

	_ = err
	return exCls, nil
}

func (c *externalClients) ServiceDiscovery() clients.ServiceDiscovery {
	return c.makosh
}

func (c *externalClients) Configurator() clients.Configurator {
	return c.configurator
}
