package e2e

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/app"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/clients/node_clients/local_state"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/middleware"
	"go.vervstack.ru/Velez/internal/transport"
)

var (
	defaultConfigPath string
	testDataDir       string
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	// filename → tests/e2e/helper_environment.go; go up two levels to reach tests/
	testsDir := filepath.Dir(filepath.Dir(filename))
	defaultConfigPath = filepath.Join(testsDir, "config_mocks", "test_config.yaml")
	testDataDir = filepath.Join(testsDir, "test_data")
}

type TestEnvironment struct {
	t          *testing.T
	configPath string

	app.App

	grpcConn *grpc.ClientConn
}

type TestEnvOpt func(a *TestEnvironment)
type StateOpt func(a *local_state.State)

func WithMatreshka() TestEnvOpt {
	return func(a *TestEnvironment) {
		a.Cfg.Environment.MatreshkaIsEnabled = true
	}
}

func WithState(t *testing.T, stateOps ...StateOpt) TestEnvOpt {
	return func(a *TestEnvironment) {
		st := readDefaultState(t)

		for _, op := range stateOps {
			op(&st)
		}

		statePath := writeState(t, st)
		a.Cfg.Environment.LocalStatePath = statePath
	}
}

func WithStateVcnEnabled() StateOpt {
	return func(a *local_state.State) {
		a.Network.Headscale.ServerUrl = "http://localhost:8080"
	}
}

func WithConfigPath(path string) TestEnvOpt {
	return func(a *TestEnvironment) {
		a.configPath = path
	}
}

func NewEnvironment(t *testing.T, opts ...TestEnvOpt) *TestEnvironment {
	t.Helper()

	var env TestEnvironment
	env.t = t
	env.App = app.App{
		Ctx: t.Context(),
		Cfg: config.Config{},
	}

	// Pre-config pass: collect opts that affect config loading (e.g. WithConfigPath).
	for _, opt := range opts {
		opt(&env)
	}

	if env.configPath == "" {
		env.configPath = defaultConfigPath
	}

	var err error

	env.App.Cfg, err = config.Load(env.configPath)
	require.NoError(t, err)

	defaultSt := readDefaultState(t)
	env.App.Cfg.Environment.LocalStatePath = writeState(t, defaultSt)

	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	env.App.ServerMaster, err = transport.NewServerManager(env.App.Ctx, lis)
	require.NoError(t, err)

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	// Post-config pass: re-apply opts that modify loaded config fields (e.g. WithMatreshka, WithState).
	for _, opt := range opts {
		opt(&env)
	}

	env.App.Cfg.AppInfo.Name = GetServiceName(t)
	env.App.Cfg.AppInfo.Version = GetServiceName(t)

	env.App.Cfg.Environment.CustomLabels = append(env.App.Cfg.Environment.CustomLabels,
		testCaseNameLabel+"="+t.Name())

	err = env.App.Custom.Init(&env.App)
	require.NoError(t, err)

	env.clean()
	t.Cleanup(env.clean)

	go func() { env.ServerMaster.Start() }()

	t.Cleanup(func() {
		e := env.ServerMaster.Stop()
		require.NoError(t, e)
	})

	env.grpcConn, err = grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			func(ctx context.Context,
				method string,
				req, reply any,
				cc *grpc.ClientConn,
				invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

				stateManager := env.Custom.NodeClients.LocalStateManager()
				localState := stateManager.Get()
				ctx = metadata.AppendToOutgoingContext(ctx, middleware.AuthHeader, localState.VelezKey)
				return invoker(ctx, method, req, reply, cc, opts...)
			}),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		closeErr := env.grpcConn.Close()
		require.NoError(t, closeErr)
	})

	return &env
}

func (e *TestEnvironment) CreateSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (*velez_api.Smerd, error) {
	if req.Labels == nil {
		req.Labels = map[string]string{}
	}

	addTestLabels(e.t, req.Labels)

	response, err := e.Custom.ApiGrpcImpl.CreateSmerd(ctx, req)
	if err != nil {
		return nil, err
	}

	removeTestLabels(response.Labels)

	return response, nil
}

func (e *TestEnvironment) ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error) {
	return e.Custom.ApiGrpcImpl.ListSmerds(ctx, req)
}

func (e *TestEnvironment) DropSmerd(ctx context.Context, req *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error) {
	return e.Custom.ApiGrpcImpl.DropSmerd(ctx, req)
}

func (e *TestEnvironment) VpnClient() velez_api.VcnApiClient {
	return velez_api.NewVcnApiClient(e.grpcConn)
}

func (e *TestEnvironment) clean() {
	ctx := context.Background()
	dockerClient := e.Custom.NodeClients.Docker().Client()

	listReq := &velez_api.ListSmerds_Request{
		Label: map[string]string{
			testCaseNameLabel: e.t.Name(),
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

func readDefaultState(t *testing.T) local_state.State {
	t.Helper()

	var defaultState local_state.State

	defaultStateFilePath := filepath.Join(testDataDir, "default-private-key.json")
	defaultStateFile, err := os.Open(defaultStateFilePath)
	require.NoError(t, err)
	defer func() {
		fErr := defaultStateFile.Close()
		require.NoError(t, fErr)
	}()

	err = json.NewDecoder(defaultStateFile).Decode(&defaultState)
	require.NoError(t, err)

	return defaultState
}

func writeState(t *testing.T, st local_state.State) (statePath string) {
	t.Helper()

	dirPath := t.TempDir()
	statePath = filepath.Join(dirPath, "state.json")

	f, err := os.Create(statePath)
	require.NoError(t, err)
	defer func() {
		fErr := f.Close()
		require.NoError(t, fErr)
	}()

	t.Cleanup(func() {
		require.NoError(t, os.Remove(statePath))
	})

	err = json.NewEncoder(f).Encode(st)
	require.NoError(t, err)

	return statePath
}
