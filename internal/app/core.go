package app

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/clients/managers"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/utils/closer"
)

func (a *App) MustInitCore() {
	var err error

	if a.Ctx == nil {
		a.Ctx = context.Background()
	}

	var cancel context.CancelFunc
	a.Ctx, cancel = context.WithCancel(a.Ctx)
	closer.Add(func() error {
		cancel()
		return nil
	})

	// Load config
	{
		a.Cfg, err = config.Load()
		if err != nil {
			logrus.Fatalf("error reading config %s", err.Error())
		}

	}

	a.Clients, err = managers.New(a.Ctx, a.Cfg)
	if err != nil {
		logrus.Fatalf("error initializing clients %s", err.Error())
	}

	return
}
