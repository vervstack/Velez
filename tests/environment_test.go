package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.vervstack.ru/matreshka/pkg/matreshka"

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
		Cfg: config.Config{
			AppInfo: matreshka.AppInfo{
				Name:            getServiceName(t),
				Version:         getServiceName(t),
				StartupDuration: 10 * time.Second,
			},
			Servers:         config.ServersConfig{},
			Environment:     config.EnvironmentConfig{},
			Overrides:       matreshka.ServiceDiscovery{},
			MatreshkaConfig: matreshka.AppConfig{},
		},
	}
	// init basic config located at ./config/*.yaml
	err := env.App.InitConfig()
	require.NoError(t, err)

	for _, opt := range opts {
		opt(&env.App)
	}

	err = env.App.Custom.Init(&env.App)
	require.NoError(t, err)

	t.Cleanup(func() {
		// TODO perform clean up
	})

	return &env
}
