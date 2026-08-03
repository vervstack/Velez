package test_helper

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
)

const (
	// LabelBasedDockerHostEnvVar lets a developer point the label-based-runtime
	// real-Docker tests at a specific daemon/socket (a different existing
	// Docker context, a remote host, or - later - an isolated instance)
	// without touching test code. Named per-implementation (not a single
	// blanket VELEZ_TEST_DOCKER_HOST) since other ContainerRuntime
	// implementations tested this way later will want their own,
	// independently overridable socket. Unset means "use today's default":
	// client.FromEnv, i.e. whatever the ambient Docker context/DOCKER_HOST
	// already resolves to.
	LabelBasedDockerHostEnvVar = "VELEZ_TEST_LABEL_BASED_DOCKER_HOST"

	// HelloWorldAppImage is the tiny fixture image real-Docker tests build
	// containers from. Duplicated from tests/e2e/helper.go:21's
	// HelloWorldAppImage constant rather than imported: tests/e2e imports
	// internal/app, which imports internal/jobs, so importing package e2e
	// from internal/jobs (a consumer of this package) would create an import
	// cycle.
	HelloWorldAppImage = "godverv/hello_world:v0.0.14"
)

func dockerClientOpts() []client.Opt {
	opts := []client.Opt{client.FromEnv, client.WithAPIVersionNegotiation()}

	if host := os.Getenv(LabelBasedDockerHostEnvVar); host != "" {
		opts = append(opts, client.WithHost(host))
	}

	return opts
}

// UniqueName mirrors tests/e2e's helper.GetServiceName sanitization,
// parameterized by prefix so callers across packages get collision-free
// names for every Docker-global-namespace resource (container names, network
// names, environment/suffix strings, ...) under t.Parallel().
func UniqueName(t *testing.T, prefix string) string {
	t.Helper()

	return prefix + "_" + strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
}

// NewRealDockerAPI returns a real Docker SDK client bound to the daemon
// resolved by dockerClientOpts (LabelBasedDockerHostEnvVar override, or the
// ambient DOCKER_HOST/context if unset) - the raw client.APIClient shape
// container_runtime's newLabelRuntime(api, suffix, bakedLabels) needs.
func NewRealDockerAPI(t *testing.T) client.APIClient {
	t.Helper()

	cli, err := client.NewClientWithOpts(dockerClientOpts()...)
	require.NoError(t, err)

	return cli
}

// NewRealDocker returns the real node_clients.Docker wrapper bound to the
// same resolved daemon as NewRealDockerAPI, for callers that need
// node_clients.Docker's higher-level shape (e.g. internal/jobs's NodeClients
// adapter).
func NewRealDocker(t *testing.T) *docker.Docker {
	t.Helper()

	d, err := docker.NewClientWithOpts(nil, dockerClientOpts()...)
	require.NoError(t, err)

	return d
}

var (
	pullOnceMu sync.Mutex
	pullOnce   = map[string]*sync.Once{}
)

// EnsurePulled pulls image at most once per test binary run, guarded by a
// sync.Once per image - mirrors ports.go's GetSharedPortManager once-pattern,
// parameterized so multiple fixture images don't share a single guard.
func EnsurePulled(t *testing.T, cli client.APIClient, image string) {
	t.Helper()

	pullOnceMu.Lock()

	once, ok := pullOnce[image]
	if !ok {
		once = &sync.Once{}
		pullOnce[image] = once
	}

	pullOnceMu.Unlock()

	once.Do(func() {
		_, pullErr := dockerutils.PullImage(t.Context(), cli, image, false)
		require.NoError(t, pullErr)
	})
}

// RemoveContainer force-removes id, ignoring "no such container" so it's
// safe to register unconditionally via t.Cleanup even when the test itself
// already removed (or never created) the container.
//
// Deliberately uses context.Background(), not t.Context(): this is meant to
// be called from a t.Cleanup callback, and testing.T cancels t.Context()
// right before running Cleanup funcs - using it here would make every
// cleanup-time removal fail with "context canceled" before ever reaching the
// daemon.
func RemoveContainer(t *testing.T, cli client.APIClient, id string) {
	t.Helper()

	if id == "" {
		return
	}

	removeOpts := container.RemoveOptions{Force: true}

	err := cli.ContainerRemove(context.Background(), id, removeOpts)
	if err != nil && !strings.Contains(err.Error(), docker.NoSuchContainerError) {
		require.NoError(t, err)
	}
}
