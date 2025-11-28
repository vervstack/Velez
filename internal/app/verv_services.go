package app

import (
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/app/matreshka_client"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"

	"go.vervstack.ru/Velez/internal/backservice/configuration"
	"go.vervstack.ru/Velez/internal/backservice/env"
	headscaleBackservice "go.vervstack.ru/Velez/internal/backservice/headscale"
	"go.vervstack.ru/Velez/internal/backservice/service_discovery"
	"go.vervstack.ru/Velez/internal/clients/headscale"
	"go.vervstack.ru/Velez/internal/clients/matreshka"
)

func (c *Custom) setupVervServices(a *App) error {
	err := c.initServiceDiscovery(a)
	if err != nil {
		return rerrors.Wrap(err, "error initializing service discovery")
	}

	err = c.initConfigurationService(a)
	if err != nil {
		return rerrors.Wrap(err, "error initializing configuration service")
	}

	err = c.initPrivateNetwork(a)
	if err != nil {
		return rerrors.Wrap(err, "error initializing private network")
	}

	return nil
}

func (c *Custom) initServiceDiscovery(a *App) (err error) {
	if a.Cfg.Environment.NodeMode {
		service_discovery.LaunchMakosh(a.Ctx, &a.Cfg, c.NodeClients)
	}

	c.ServiceDiscovery, err = service_discovery.SetupServiceDiscovery(
		a.Cfg.Environment.MakoshURL,
		a.Cfg.Environment.MakoshKey,
		a.Cfg.Overrides,
	)
	if err != nil {
		return rerrors.Wrap(err, "error initializing service discovery ")
	}

	return nil
}

func (c *Custom) initConfigurationService(a *App) (err error) {
	// TODO change onto granular enabler for each verv service
	if a.Cfg.Environment.NodeMode {
		configuration.LaunchMatreshka(a.Ctx, &a.Cfg, c.NodeClients, c.ServiceDiscovery)
	}

	c.MatreshkaClient, err = matreshka.NewClient(
		grpc.WithUnaryInterceptor(
			matreshka_client.WithHeader(
				matreshka_client.Pass, c.NodeClients.LocalStateManager().Get().MatreshkaKey)))
	if err != nil {
		return rerrors.Wrap(err, "error creating matreshka grpc client")
	}

	apiVersion, err := c.MatreshkaClient.ApiVersion(a.Ctx, &matreshka_api.ApiVersion_Request{})
	if err != nil {
		return rerrors.Wrap(err, "can't ping matreshka api")
	}
	_ = apiVersion

	storeConfig := &matreshka_api.StoreConfig_Request{
		Format:     matreshka_api.Format_yaml,
		ConfigName: env.VelezName,
		Config:     nil,
	}

	storeConfig.Config, err = yaml.Marshal(a.Cfg.MatreshkaConfig)
	if err != nil {
		return rerrors.Wrap(err, "error marshalling original config before saving it to matreshka")
	}

	_, err = c.MatreshkaClient.StoreConfig(a.Ctx, storeConfig)
	if err != nil {
		return rerrors.Wrap(err, "error storing config in matreshka")
	}

	return nil
}

func (c *Custom) initPrivateNetwork(a *App) (err error) {
	if a.Cfg.Environment.VpnIsEnabled {
		headscaleBackservice.Launch(a.Ctx, a.Cfg, c.NodeClients)
	}

	c.HeadscaleClient, err = headscale.New(a.Ctx, c.NodeClients, headscaleBackservice.Name)
	if err != nil {
		return rerrors.Wrap(err, "")
	}

	return nil
}
