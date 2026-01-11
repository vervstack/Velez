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

// TestEnvironment exposes only what is needed in test suite (API client, other node and cluster clients with test wrappers)
// TODO as for now TestEnvironment is fully exposed. It suppose to be the other way around.
type TestEnvironment struct {
	app.App

	t *testing.T
}

type TestEnvOpt func(a *TestEnvironment)

func WithMatreshka() TestEnvOpt {
	return func(a *TestEnvironment) {
		a.Cfg.Environment.MatreshkaIsEnabled = true
	}
}

func WithVcn() TestEnvOpt {
	return func(a *TestEnvironment) {
		a.Cfg.Environment.CustomPassToKey
	}
}

func NewEnvironment(t *testing.T, opts ...TestEnvOpt) *TestEnvironment {
	var env TestEnvironment

	env.t = t

	env.App = app.App{
		Ctx: t.Context(),
		Cfg: config.Config{},
	}

	var err error

	//region Default Test Application setup
	// Bare minimal functionality. Just Api for containers

	// init basic config located at ./config/*.yaml
	env.App.Cfg, err = config.Load("./config_mocks/test_config.yaml")
	require.NoError(t, err)

	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	env.App.ServerMaster, err = transport.NewServerManager(env.App.Ctx, lis)
	require.NoError(t, err)

	//endregion

	for _, opt := range opts {
		opt(&env)
	}

	env.App.Cfg.AppInfo.Name = GetServiceName(t)
	env.App.Cfg.AppInfo.Version = GetServiceName(t)

	//region Application start
	err = env.App.Custom.Init(&env.App)
	require.NoError(t, err)
	//endregion
	env.clean()
	t.Cleanup(env.clean)

	return &env
}

func (e *TestEnvironment) CreateSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (smerd *velez_api.Smerd, err error) {
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

func (e *TestEnvironment) clean() {
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
