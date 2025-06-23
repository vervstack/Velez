package app

import (
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/app/matreshka_client"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/backservice/configuration"
	"go.vervstack.ru/Velez/internal/backservice/env"
	"go.vervstack.ru/Velez/internal/backservice/service_discovery"
	"go.vervstack.ru/Velez/internal/clients/matreshka"
)

func (c *Custom) setupVervNodeEnvironment() (err error) {
	// Verv network for communication inside node
	err = env.StartNetwork(c.NodeClients.Docker())
	if err != nil {
		return rerrors.Wrap(err, "error creating network")
	}

	// Verv volumes for persistence inside node
	err = env.StartVolumes(c.NodeClients.Docker())
	if err != nil {
		return rerrors.Wrap(err, "error creating volumes")
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
	if a.Cfg.Environment.NodeMode {
		configuration.LaunchMatreshka(a.Ctx, &a.Cfg, c.NodeClients, c.ServiceDiscovery)
	}

	c.MatreshkaClient, err = matreshka.NewClient(
		grpc.WithUnaryInterceptor(
			matreshka_client.WithHeader(
				matreshka_client.Pass, a.Cfg.Environment.MatreshkaKey)))
	if err != nil {
		return rerrors.Wrap(err, "error creating matreshka grpc client")
	}

	_, err = c.MatreshkaClient.ApiVersion(a.Ctx, &matreshka_api.ApiVersion_Request{})
	if err != nil {
		return rerrors.Wrap(err, "can't ping matreshka api")
	}

	return nil
}
