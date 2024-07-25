package app

import (
	"github.com/godverv/Velez/internal/backservice/portainer"
	"github.com/godverv/Velez/internal/backservice/watchtower"
	"github.com/godverv/Velez/internal/cron"
)

func (a *App) initBackServices() {
	if a.Cfg.GetEnvironment().WatchTowerEnabled {
		go cron.KeepAlive(a.Ctx, watchtower.New(a.Cfg, a.Services.GetContainerManagerService()))
	}

	if a.Cfg.GetEnvironment().PortainerEnabled {
		go cron.KeepAlive(a.Ctx, portainer.New(a.Services.GetContainerManagerService()))
	}
}
