package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/app"
	"go.vervstack.ru/Velez/internal/config"
)

type Environment struct {
	app.App
}

type envOpt func(a *app.App)

func WithFullConfig(cfg config.Config) envOpt {
	return func(a *app.App) {
		a.Cfg = cfg
	}
}

func NewEnvironment(t *testing.T, opts ...envOpt) *Environment {
	var env Environment

	env.App = app.App{
		Ctx: t.Context(),
		Cfg: config.Config{},
	}
	// init basic config located at ./config/*.yaml
	var err error
	env.App.Cfg, err = config.Load("./config/config.yaml")
	require.NoError(t, err)

	for _, opt := range opts {
		opt(&env.App)
	}

	env.App.Cfg.AppInfo.Name = getServiceName(t)
	env.App.Cfg.AppInfo.Version = getServiceName(t)

	err = env.App.Custom.Init(&env.App)
	require.NoError(t, err)

	t.Cleanup(func() {
		// TODO perform clean up
	})

	return &env
}
