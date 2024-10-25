package app

import (
	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/makosh/pkg/makosh_be"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/backservice/configuration"
	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/backservice/service_discovery"
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

func (c *Custom) initServiceDiscovery(a *App) {
	sdConn := service_discovery.ServiceDiscoveryConnection{
		Addr:  a.Cfg.Environment.MakoshUrls,
		Token: a.Cfg.Environment.MakoshKey,
	}

	if a.Cfg.Environment.NodeMode {
		sdConn = service_discovery.InitInstance(a.Ctx, a.Cfg, c.NodeClients)
	} else {
		// TODO add multiple makosh urls handling

	}

	_, err := service_discovery.NewServiceDiscovery(sdConn.Addr[0], sdConn.Token)
	if err != nil {
		logrus.Fatalf("error initializing service discovery %s", err)
	}
}

func (c *Custom) initConfigurationService(a *App) {
	var matreshkaConn configuration.MatreshkaConnect

	if a.Cfg.Environment.NodeMode {
		matreshkaConn = configuration.InitInstance(a.Ctx, a.Cfg, c.NodeClients)
	} else {
		// TODO add multiple matreshka urls handling
		matreshkaConn.Addr = a.Cfg.Environment.MatreshkaUrls[0]
	}

	matreshkaEndpoints := &makosh_be.UpsertEndpoints_Request{
		Endpoints: []*makosh_be.Endpoint{
			{
				ServiceName: "matreshka",
				Addrs:       []string{matreshkaConn.Addr},
			},
		},
	}
	_ = matreshkaEndpoints
	//_, err := c.ServiceDiscovery.Api.UpsertEndpoints(a.Ctx, matreshkaEndpoints)
	//if err != nil {
	//logrus.Fatalf("error upserting endpoints for matreshka: %s", err)
	//}

	println(1)
}
