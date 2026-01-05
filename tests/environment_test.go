package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/test/bufconn"

	"go.vervstack.ru/Velez/internal/app"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/transport"
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

	var err error

	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	env.App.ServerMaster, err = transport.NewServerManager(env.App.Ctx, lis)
	require.NoError(t, err)

	// init basic config located at ./config/*.yaml
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
