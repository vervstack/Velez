package tests

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/test/bufconn"

	"go.vervstack.ru/Velez/internal/app"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/transport"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

// Environment exposes only what is needed in test suite (API client, other node and cluster clients with test wrappers)
// TODO as for now Environment is fully exposed. It suppose to be the other way around.
type Environment struct {
	app.App

	t *testing.T
}

type EnvOpt func(a *app.App)

func WithMatreshka() EnvOpt {
	return func(a *app.App) {
		a.Cfg.Environment.MatreshkaIsEnabled = true
	}
}

func NewEnvironment(t *testing.T, opts ...EnvOpt) *Environment {
	var env Environment

	env.t = t

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
	env.App.Cfg, err = config.Load("./config_mocks/test_config.yaml")
	require.NoError(t, err)
	for _, opt := range opts {
		opt(&env.App)
	}

	env.App.Cfg.AppInfo.Name = GetServiceName(t)
	env.App.Cfg.AppInfo.Version = GetServiceName(t)

	err = env.App.Custom.Init(&env.App)
	require.NoError(t, err)

	env.clean()
	t.Cleanup(env.clean)

	return &env
}

func (e *Environment) CreateSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (smerd *velez_api.Smerd, err error) {
	if req.Labels == nil {
		req.Labels = map[string]string{}
	}

	addTestLabels(e.t, req.Labels)

	response, err := e.App.Custom.ApiGrpcImpl.CreateSmerd(ctx, req)
	if err != nil {
		return nil, err
	}

	removeTestLabels(response.Labels)

	return response, nil
}

func (e *Environment) clean() {
	ctx := context.Background()
	dockerClient := e.Custom.NodeClients.Docker().Client()

	listReq := &velez_api.ListSmerds_Request{
		Label: map[string]string{
			integrationTestLabel: "true",
			testCaseNameLabel:    e.t.Name(),
		},
	}

	cList, err := dockerutils.ListContainers(ctx, dockerClient, listReq)
	if err != nil {
		logrus.Fatal(err)
	}

	for _, cont := range cList {
		removeOps := container.RemoveOptions{
			Force: true,
		}

		err = dockerClient.ContainerRemove(ctx, cont.ID, removeOps)
		if err != nil {
			logrus.Fatal(err)
		}
	}
}
