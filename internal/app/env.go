package app

import (
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/backservice/configuration"
	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/cron"
)

func (a *App) MustInitEnvironment() {
	err := env.StartNetwork(a.Docker)
	if err != nil {
		logrus.Fatalf("error creating network: %s", err)
	}

	err = env.StartVolumes(a.Docker)
	if err != nil {
		logrus.Fatalf("error creating volumes %s", err)
	}

	if !a.Cfg.GetEnvironment().NodeMode {
		return
	}

	var portToExposeTo string
	if a.Cfg.GetEnvironment().ExposeMatreshkaPort {
		p := uint64(a.Cfg.GetEnvironment().MatreshkaPort)

		if p == 0 {
			portFromPool, err := a.PortManager.GetPort()
			if err != nil {
				logrus.Fatalf("no available port for config to expose")
				return
			}

			p = uint64(portFromPool)
		}

		portToExposeTo = strconv.FormatUint(p, 10)
	}

	conf := configuration.New(a.Docker, portToExposeTo)
	err = conf.Start()
	if err != nil {
		logrus.Fatalf("error launching config backservice: %s", err)
	}

	go cron.KeepAlive(a.Ctx, conf)
}
