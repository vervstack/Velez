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
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/app"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/clients/node_clients/local_state"
	"go.vervstack.ru/Velez/internal/cluster/configuration"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/middleware"
	"go.vervstack.ru/Velez/tests/test_helper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

// sharedPortManagerImpl is a single PortManager shared across all parallel test environments.
// Initialized once on first NewEnvironment call.

var (
	defaultConfigPath string
	testDataDir       string
)

func init() {
	pc, filename, _, _ := runtime.Caller(0)
	_ = pc
	// filename → tests/e2e/helper_environment.go; go up two levels to reach tests/
	testsDir := filepath.Dir(filepath.Dir(filename))
	defaultConfigPath = filepath.Join(testsDir, "config_mocks", "velez_default_config.yaml")
	testDataDir = filepath.Join(testsDir, "test_data")
}

type TestEnvironment struct {
	t          *testing.T
	configPath string

	app.App

	grpcConn *grpc.ClientConn
}

type (
	TestEnvOpt func(a *TestEnvironment)
	StateOpt   func(a *local_state.State)
)

// WithMatreshka enables cluster mode against a shared matreshka container.
// The container + its keep-alive loop are created exactly once per test
// binary (see main_test.go) and reused by every TestEnvironment that calls
// this option — matreshka's container name and gRPC port are fixed by
// design (mirrors production: exactly one matreshka instance per node), so
// running more than one real container under that name races on
// create/kill. See docs/plans/e2e_flaky_lifecycle_matreshka.md.
//
// IMPORTANT: this fixture is a single-process (in-memory) singleton. Go
// compiles one test binary per package, so it only coordinates tests
// within *this* package (tests/e2e). A test in a different package runs in
// its own OS process, never observes this singleton, and silently
// reintroduces the container-collision hazard this fixture exists to
// prevent. Any test that needs WithMatreshka() must live in package e2e.
func WithMatreshka() TestEnvOpt {
	return func(a *TestEnvironment) {
		a.Cfg.Environment.MatreshkaIsEnabled = true
		a.Ctx = configuration.WithSharedInstance(a.Ctx, getSharedMatreshka(a.t))
	}
}

func WithState(t *testing.T, stateOps ...StateOpt) TestEnvOpt {
	t.Helper()

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

func WithContainerSuffix(suffix string) TestEnvOpt {
	return func(a *TestEnvironment) {
		a.Cfg.Environment.ContainerSuffix = suffix
	}
}

func WithEnvironments(envs []string) TestEnvOpt {
	return func(a *TestEnvironment) {
		a.Cfg.Environment.Environments = envs
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

	initConfig(t, &env)

	// Post-config pass: re-apply opts that modify loaded config fields (e.g. WithMatreshka, WithState).
	for _, opt := range opts {
		opt(&env)
	}

	env.Cfg.AppInfo.Name = GetServiceName(t)
	env.Cfg.AppInfo.Version = GetServiceName(t)

	env.Cfg.Environment.CustomLabels = append(
		env.Cfg.Environment.CustomLabels,
		testCaseNameLabel+"="+t.Name())

	initGrpc(t, &env)

	err := env.Custom.Init(&env.App)
	require.NoError(t, err)

	go func() {
		startServerMasterErr := env.Custom.Start(env.Ctx)
		require.NoError(t, startServerMasterErr)
	}()

	t.Cleanup(func() {
		e := env.Custom.Stop()
		require.NoError(t, e)
	})

	portManager := test_helper.GetSharedPortManager(t, env.Cfg.Environment.AvailablePorts)
	env.Custom.NodeClients.PortManagerContainer().
		Set(portManager)

	// Cleaning dished before and after dinner just in case
	env.clean()
	t.Cleanup(env.clean)

	return &env
}

func initConfig(t *testing.T, env *TestEnvironment) {
	t.Helper()

	if env.configPath == "" {
		env.configPath = defaultConfigPath
	}

	var err error

	env.Cfg, err = config.Load(env.configPath)
	require.NoError(t, err)

	defaultSt := readDefaultState(t)
	env.Cfg.Environment.LocalStatePath = writeState(t, defaultSt)
}

func initGrpc(t *testing.T, env *TestEnvironment) {
	t.Helper()

	const bufSize = 1024 * 1024

	lis := bufconn.Listen(bufSize)

	env.MASTER = lis

	var err error

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	env.grpcConn, err = grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			func(ctx context.Context,
				method string,
				req, reply any,
				cc *grpc.ClientConn,
				invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
			) error {
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
}

func (e *TestEnvironment) CreateSmerd(t *testing.T, req *velez_api.CreateSmerd_Request) *velez_api.Smerd {
	t.Helper()

	ctx := t.Context()

	if req.Labels == nil {
		req.Labels = map[string]string{}
	}

	addTestLabels(e.t, req.Labels)

	response, err := e.Custom.ApiGrpcImpl.CreateSmerd(ctx, req)
	require.NoError(t, err)

	removeTestLabels(response.Labels)

	return response
}

func (e *TestEnvironment) ListSmerds(t *testing.T, ctx context.Context, req *velez_api.ListSmerds_Request) *velez_api.ListSmerds_Response {
	t.Helper()

	resp, err := e.Custom.ApiGrpcImpl.ListSmerds(ctx, req)
	require.NoError(t, err)

	return resp
}

func (e *TestEnvironment) DropSmerd(
	ctx context.Context,
	t *testing.T,
	req *velez_api.DropSmerd_Request,
) *velez_api.DropSmerd_Response {
	t.Helper()

	resp, err := e.Custom.ApiGrpcImpl.DropSmerd(ctx, req)
	require.NoError(t, err)

	return resp
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
		log.Fatal().Err(err).Send()
	}

	for _, cont := range cList {
		removeOps := container.RemoveOptions{
			Force: true,
		}

		err = dockerClient.ContainerRemove(ctx, cont.ID, removeOps)
		if err != nil {
			log.Fatal().Err(err).Send()
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
