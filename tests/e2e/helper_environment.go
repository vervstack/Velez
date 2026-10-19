package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"go.vervstack.ru/matreshka/pkg/matreshka/service_discovery"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/matreshka"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/clients/node_clients/local_state"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/middleware"
	"go.vervstack.ru/Velez/tests/dind"
)

const (
	awaitTaskTimeout = 8 * time.Minute
)

var (
	defaultConfigPath string
	testDataDir       string
)

//nolint:gochecknoinits
func init() {
	pc, filename, _, _ := runtime.Caller(0)

	_ = pc
	// filename → tests/e2e/helper_environment.go; go up two levels to reach tests/
	testsDir := filepath.Dir(filepath.Dir(filename))

	defaultConfigPath = filepath.Join(testsDir, "config_mocks", "velez_default_config.yaml")
	testDataDir = filepath.Join(testsDir, "test_data")
}

type TestEnvironment struct {
	t *testing.T

	Cfg config.Config

	grpcConn *grpc.ClientConn

	// velezKey is the auth key from the state file this environment points
	// Velez at; the test client sends it on every call.
	velezKey string

	dockerClient   *client.Client
	dockerClientMu sync.Mutex

	// clusterPgPort is the leased host port the cluster-pg sidecar is pinned
	// to; set by newStatefullEnvironment, zero for an environment without one.
	clusterPgPort uint64 //nolint:unused // read only by the e2e_full statefull helpers

	// velez is the Velez container this environment runs against.
	velez *velezContainer

	// managed is the daemon this environment's Velez manages when it is not
	// the DinD the Velez container itself lives in; nil means the same DinD.
	managed *managedDaemon
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
		getSharedMatreshka(a.t)

		a.Cfg.Environment.MatreshkaIsEnabled = true

		image, err := matreshkaImage()
		require.NoError(a.t, err)

		a.Cfg.Environment.MatreshkaImage = image

		override := &service_discovery.Override{
			ServiceName: matreshka.ServiceName,
			Urls:        service_discovery.Urls{dindGatewayAddr(a.t, dindMatreshkaPort)},
		}

		a.Cfg.Overrides.Overrides = append(a.Cfg.Overrides.Overrides, override)
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
		a.velezKey = st.VelezKey
	}
}

// WithStateVcnEnabled points the app's verv-closed-network client at an
// already-running headscale (serverURL + apiKey). Both non-empty makes
// verv_closed_network.SetupVcn take the headscale.Connect(url, key) branch -
// it connects to the given server and never launches its own headscale
// container. Pass the shared fixture's address: getSharedHeadscale(t).apiURL
// / .apiKey. loginServerURL is the address tailnet sidecars dial from inside
// the DinD (getSharedHeadscale(t).loginURL).
func WithStateVcnEnabled(serverURL, apiKey, loginServerURL string) StateOpt {
	return func(a *local_state.State) {
		a.Network.Headscale.ServerUrl = serverURL
		a.Network.Headscale.Key = apiKey
		a.Network.Headscale.LoginServerUrl = loginServerURL
	}
}

func WithContainerSuffix(suffix string) TestEnvOpt {
	return func(a *TestEnvironment) {
		a.Cfg.Environment.ContainerSuffix = suffix
	}
}

// WithClusterPgDsn injects the ClusterPgDsn advertise-address override
// (internal/config EnvironmentConfig.ClusterPgDsn) into the loaded config.
// The enable_statefull job's getRootDsnJob reads Host+Port back out of it so
// the Velez container can reach the cluster postgres sidecar running
// inside the DinD. Applied in the post-config pass, mirroring
// WithContainerSuffix.
func WithClusterPgDsn(dsn string) TestEnvOpt {
	return func(a *TestEnvironment) {
		a.Cfg.Environment.ClusterPgDsn = dsn
	}
}

// WithMigrationsDir injects the MigrationsDir override (internal/config
// EnvironmentConfig.MigrationsDir) into the loaded config, so
// sqldb.RollMigration resolves goose migrations from an absolute path
// instead of "./migrations" relative to the process's working directory -
// e2e tests don't run from the repo root. Applied in the post-config pass,
// mirroring WithContainerSuffix/WithClusterPgDsn.
func WithMigrationsDir(dir string) TestEnvOpt {
	return func(a *TestEnvironment) {
		a.Cfg.Environment.MigrationsDir = dir
	}
}

// WithAutoUpdate turns the auto-update worker on for this environment. The
// default mock config leaves it off: every environment's worker scans the
// whole shared DinD, so parallel environments would race to upgrade the same
// labelled container. Applied in the post-config pass, mirroring
// WithContainerSuffix.
func WithAutoUpdate() TestEnvOpt {
	return func(a *TestEnvironment) {
		a.Cfg.Environment.AutoUpdateIsEnabled = true
	}
}

func WithEnvironments(envs []string) TestEnvOpt {
	return func(a *TestEnvironment) {
		a.Cfg.Environment.Environments = envs
	}
}

func NewEnvironment(t *testing.T, opts ...TestEnvOpt) *TestEnvironment {
	t.Helper()

	requireDindHarness(t)

	env := TestEnvironment{t: t}

	initConfig(t, &env)

	pinDindPorts(&env)

	for _, opt := range opts {
		opt(&env)
	}

	env.Cfg.Environment.CustomLabels = append(
		env.Cfg.Environment.CustomLabels,
		testCaseNameLabel+"="+t.Name())

	startContainerBackend(t, &env)

	return &env
}

// requireDindHarness fails the test unless it is running under the DinD
// harness set up by TestMain (see tests/dind). It guards against pointing
// the suite at a developer's real Docker daemon.
func requireDindHarness(t *testing.T) {
	t.Helper()

	if os.Getenv(dind.EnvActive) == "" {
		t.Fatal("e2e suite must run under the DinD harness set up by TestMain " +
			"(tests/e2e/main_test.go); " + dind.EnvActive + " is not set")
	}
}

// pinDindPorts pins matreshka's gRPC bind to the fixed port the shared
// fixture listens on inside the DinD.
func pinDindPorts(env *TestEnvironment) {
	env.Cfg.Environment.MatreshkaPort = dindMatreshkaPort
}

func initConfig(t *testing.T, env *TestEnvironment) {
	t.Helper()

	var err error

	env.Cfg, err = config.Load(defaultConfigPath)
	require.NoError(t, err)

	defaultSt := readDefaultState(t)

	env.Cfg.Environment.LocalStatePath = writeState(t, defaultSt)
	env.velezKey = defaultSt.VelezKey
}

func connectGrpc(t *testing.T, env *TestEnvironment, target string, dialOpts ...grpc.DialOption) {
	t.Helper()

	dialOpts = append(dialOpts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			func(ctx context.Context,
				method string,
				req, reply any,
				cc *grpc.ClientConn,
				invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
			) error {
				ctx = metadata.AppendToOutgoingContext(ctx, middleware.AuthHeader, env.velezKey)

				return invoker(ctx, method, req, reply, cc, opts...)
			}),
	)

	var err error

	env.grpcConn, err = grpc.NewClient(target, dialOpts...)
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

	addTestLabels(e.t, req.GetLabels())

	response, err := e.VelezApiClient().CreateSmerd(ctx, req)
	require.NoError(t, err)

	removeTestLabels(response.GetLabels())

	return response
}

func (e *TestEnvironment) ListSmerds(
	t *testing.T,
	ctx context.Context,
	req *velez_api.ListSmerds_Request,
) *velez_api.ListSmerds_Response {
	t.Helper()

	resp, err := e.VelezApiClient().ListSmerds(ctx, req)
	require.NoError(t, err)

	return resp
}

func (e *TestEnvironment) DropSmerd(
	ctx context.Context,
	t *testing.T,
	req *velez_api.DropSmerd_Request,
) *velez_api.DropSmerd_Response {
	t.Helper()

	resp, err := e.VelezApiClient().DropSmerd(ctx, req)
	require.NoError(t, err)

	return resp
}

func (e *TestEnvironment) VpnClient() velez_api.VcnApiClient {
	return velez_api.NewVcnApiClient(e.grpcConn)
}

func (e *TestEnvironment) NetworkClient() velez_api.NetworkAPIClient {
	return velez_api.NewNetworkAPIClient(e.grpcConn)
}

func (e *TestEnvironment) ServiceApiClient() velez_api.ServiceApiClient {
	return velez_api.NewServiceApiClient(e.grpcConn)
}

func (e *TestEnvironment) TasksApiClient() velez_api.TasksApiClient {
	return velez_api.NewTasksApiClient(e.grpcConn)
}

func (e *TestEnvironment) VelezApiClient() velez_api.VelezAPIClient {
	return velez_api.NewVelezAPIClient(e.grpcConn)
}

func (e *TestEnvironment) S3Client() velez_api.S3APIClient {
	return velez_api.NewS3APIClient(e.grpcConn)
}

func (e *TestEnvironment) PgaasClient() velez_api.PostgresAPIClient {
	return velez_api.NewPostgresAPIClient(e.grpcConn)
}

func (e *TestEnvironment) ContainerRegistryClient() velez_api.ContainerRegistryAPIClient {
	return velez_api.NewContainerRegistryAPIClient(e.grpcConn)
}

func (e *TestEnvironment) ControlPlaneClient() velez_api.ControlPlaneAPIClient {
	return velez_api.NewControlPlaneAPIClient(e.grpcConn)
}

func (e *TestEnvironment) SettingsClient() velez_api.SettingsAPIClient {
	return velez_api.NewSettingsAPIClient(e.grpcConn)
}

func (e *TestEnvironment) AutoUpdateClient() velez_api.AutoUpdateAPIClient {
	return velez_api.NewAutoUpdateAPIClient(e.grpcConn)
}

func (e *TestEnvironment) RunnersClient() velez_api.RunnersAPIClient {
	return velez_api.NewRunnersAPIClient(e.grpcConn)
}

// DockerClient is a test-owned client to the daemon this environment runs
// against, independent of the app under test.
func (e *TestEnvironment) DockerClient() *client.Client {
	e.dockerClientMu.Lock()
	defer e.dockerClientMu.Unlock()

	if e.dockerClient != nil {
		return e.dockerClient
	}

	dockerClient, err := client.NewClientWithOpts(
		client.WithHost(sharedDind.DockerHost), client.WithAPIVersionNegotiation())
	require.NoError(e.t, err)

	e.t.Cleanup(func() {
		closeErr := dockerClient.Close()
		require.NoError(e.t, closeErr)
	})

	e.dockerClient = dockerClient

	return dockerClient
}

// ManagedDockerClient is a test-owned client to the daemon this environment's
// Velez manages: the managed daemon when one is set, else DockerClient.
func (e *TestEnvironment) ManagedDockerClient() *client.Client {
	if e.managed == nil {
		return e.DockerClient()
	}

	return e.managed.client
}

// AwaitTask follows the task through WatchTask until it reaches DONE or FAILED
// and returns its final status.
func (e *TestEnvironment) AwaitTask(t *testing.T, entityId string, action string) *velez_api.TaskStatus {
	t.Helper()

	req := &velez_api.WatchTask_Request{
		EntityId: entityId,
		Action:   action,
	}

	ctx, cancel := context.WithTimeout(t.Context(), awaitTaskTimeout)
	defer cancel()

	stream, err := e.TasksApiClient().WatchTask(ctx, req)
	require.NoError(t, err)

	finalStatus := &velez_api.TaskStatus{}

	for {
		status, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			return finalStatus
		}

		require.NoError(t, recvErr)

		finalStatus = status
	}
}

func (e *TestEnvironment) clean() {
	cleaned := []*client.Client{e.DockerClient()}

	managed := e.ManagedDockerClient()
	if managed != cleaned[0] {
		cleaned = append(cleaned, managed)
	}

	for _, dockerClient := range cleaned {
		removeLabelledContainers(dockerClient, e.t.Name())
	}
}

func removeLabelledContainers(dockerClient *client.Client, testName string) {
	ctx := context.Background()

	listReq := &velez_api.ListSmerds_Request{
		Label: map[string]string{
			testCaseNameLabel: testName,
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
