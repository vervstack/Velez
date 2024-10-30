package app

import (
	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/makosh/pkg/makosh_be"

	"github.com/godverv/Velez/internal/backservice/configuration"
	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/backservice/service_discovery"
	"github.com/godverv/Velez/internal/clients/matreshka"
)

func (c *Custom) setupVervNodeEnvironment() (err error) {
	// Verv network for communication inside node
	err = env.StartNetwork(c.NodeClients.Docker())
	if err != nil {
		return errors.Wrap(err, "error creating network")
	}

	// Verv volumes for persistence inside node
	err = env.StartVolumes(c.NodeClients.Docker())
	if err != nil {
		return errors.Wrap(err, "error creating volumes")
	}

	return nil
}

func (c *Custom) initServiceDiscovery(a *App) (err error) {
	if a.Cfg.Environment.NodeMode {
		service_discovery.LaunchMakosh(a.Ctx, &a.Cfg, c.NodeClients)
	}

	c.MakoshClient, err = service_discovery.SetupServiceDiscovery(
		a.Cfg.Environment.MakoshURL,
		a.Cfg.Environment.MakoshKey,
	)
	if err != nil {
		return errors.Wrap(err, "error initializing service discovery ")
	}

	return nil
}

func (c *Custom) initConfigurationService(a *App) (err error) {
	matreshkaEndpoints := &makosh_be.UpsertEndpoints_Request{
		Endpoints: []*makosh_be.Endpoint{
			{
				ServiceName: configuration.Name,
				Addrs:       make([]string, 1),
			},
		},
	}

	if a.Cfg.Environment.NodeMode {
		conn := configuration.LaunchMatreshka(a.Ctx, a.Cfg, c.NodeClients)
		matreshkaEndpoints.Endpoints[0].Addrs[0] = conn.Addr
	} else if a.Cfg.Environment.MatreshkaURL != "verv://matreshka" {
		matreshkaEndpoints.Endpoints[0].Addrs[0] = a.Cfg.Environment.MatreshkaURL
	} else {
		return nil
	}

	_, err = c.MakoshClient.UpsertEndpoints(a.Ctx, matreshkaEndpoints)
	if err != nil {
		return errors.Wrap(err, "error upserting endpoints for matreshka")
	}

	c.MatreshkaClient, err = matreshka.NewClient()
	if err != nil {
		return errors.Wrap(err, "error creating matreshka grpc client")
	}

	return nil
}
