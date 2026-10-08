//go:build e2e_full

package e2e

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/gitlab_runner_config"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	buildkitImage       = "moby/buildkit:v0.34.0"
	buildkitProbeTick   = 2 * time.Second
	buildkitProbeWindow = 2 * time.Minute

	tunnelProbeNetwork = "e2e-tunnel-probe"
)

func newSetRunnerBuildkitRequest(name string, isEnabled bool) *velez_api.SetRunnerBuildkit_Request {
	return &velez_api.SetRunnerBuildkit_Request{Name: name, IsBuildkitEnabled: isEnabled}
}

// nodeResolver is the resolver the product uses, built over the test node's daemon.
func nodeResolver(env *TestEnvironment) container_runtime.RuntimeResolver {
	docker := env.Custom.NodeClients.Docker()

	return container_runtime.NewResolver(docker.Client(), docker.Host(), nil, nil)
}

// execInDind runs the docker CLI inside the DinD container, against the daemon
// that lives there - not through the tunnel under test. Output is demultiplexed
// here because the product runtime keeps the stream framing bytes.
func execInDind(t *testing.T, env *TestEnvironment, dindName string, args ...string) (string, int) {
	t.Helper()

	nodeClient := env.Custom.NodeClients.Docker().Client()

	execOpts := container.ExecOptions{
		Cmd:          append([]string{"docker"}, args...),
		AttachStdout: true,
		AttachStderr: true,
	}

	created, err := nodeClient.ContainerExecCreate(t.Context(), dindName, execOpts)
	require.NoError(t, err)

	attached, err := nodeClient.ContainerExecAttach(t.Context(), created.ID, container.ExecAttachOptions{})
	require.NoError(t, err)

	defer attached.Close()

	var stdout, stderr bytes.Buffer

	_, err = stdcopy.StdCopy(&stdout, &stderr, attached.Reader)
	require.NoError(t, err)

	inspected, err := nodeClient.ContainerExecInspect(t.Context(), created.ID)
	require.NoError(t, err)

	return strings.TrimSpace(stdout.String() + stderr.String()), inspected.ExitCode
}

func innerDockerLines(t *testing.T, env *TestEnvironment, dindName string, args ...string) []string {
	t.Helper()

	output, exitCode := execInDind(t, env, dindName, args...)
	require.Zero(t, exitCode, output)

	if output == "" {
		return nil
	}

	return strings.Split(output, "\n")
}

// preloadBuildkitImage copies buildkitImage from the test daemon into the
// runner DinD over the tunnel, so the enable flow never depends on registry
// download speed. The closer NestedRuntime returns is the tunnel docker client.
func preloadBuildkitImage(t *testing.T, env *TestEnvironment, dindName string) {
	t.Helper()

	nodeClient := env.Custom.NodeClients.Docker().Client()

	pulled, err := nodeClient.ImagePull(t.Context(), buildkitImage, image.PullOptions{})
	require.NoError(t, err)

	_, err = io.Copy(io.Discard, pulled)
	require.NoError(t, err)
	require.NoError(t, pulled.Close())

	nested, closer, err := nodeResolver(env).NestedRuntime(t.Context(), "", dindName)
	require.NoError(t, err)

	t.Cleanup(func() { _ = closer.Close() })

	require.Eventually(t, func() bool {
		_, listErr := nested.ListNetworks(t.Context(), true)

		return listErr == nil
	}, buildkitProbeWindow, buildkitProbeTick, "inner daemon never answered through the tunnel")

	tunnel, isClient := closer.(client.APIClient)
	require.True(t, isClient, "the nested closer must be the tunnel docker client")

	archive, err := nodeClient.ImageSave(t.Context(), []string{buildkitImage})
	require.NoError(t, err)

	defer func() { _ = archive.Close() }()

	loaded, err := tunnel.ImageLoad(t.Context(), archive, client.ImageLoadWithQuiet(true))
	require.NoError(t, err)

	_, err = io.Copy(io.Discard, loaded.Body)
	require.NoError(t, err)
	require.NoError(t, loaded.Body.Close())
}

func setRunnerBuildkitAndAwait(t *testing.T, env *TestEnvironment, runnerName string, isEnabled bool) {
	t.Helper()

	runnersClient := velez_api.NewRunnersAPIClient(env.grpcConn)

	_, err := runnersClient.SetRunnerBuildkit(t.Context(), newSetRunnerBuildkitRequest(runnerName, isEnabled))
	require.NoError(t, err)

	task := awaitJobTask(t, env, runnerName, jobs.SetRunnerBuildkitAction)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "set runner buildkit task error: %s", task.Error.String)
}

func logDindOnFailure(t *testing.T, env *TestEnvironment, dindName, runnerContainer string) {
	t.Helper()

	t.Cleanup(func() {
		if !t.Failed() {
			return
		}

		output, _ := execInDind(t, env, dindName, "ps", "-a")
		t.Logf("containers inside %s:\n%s", dindName, output)

		output, _ = execInDind(t, env, dindName, "logs", domain.RunnerBuildkitServiceName(runnerContainer))
		t.Logf("buildkitd logs:\n%s", output)
	})
}

func requireRunnerNetworkMode(t *testing.T, env *TestEnvironment, runnerContainer, want string) {
	t.Helper()

	nodeRuntime, err := nodeResolver(env).Runtime(t.Context(), "")
	require.NoError(t, err)

	config, err := nodeRuntime.CopyFromContainer(t.Context(), runnerContainer, gitlab_runner_config.ConfigPath)
	require.NoError(t, err)

	got, err := gitlab_runner_config.NetworkMode(config)
	require.NoError(t, err)
	require.Equal(t, want, got, "config.toml:\n%s", config)
}

func requireBuildkitInsideDind(t *testing.T, env *TestEnvironment, dindName, runnerContainer string) {
	t.Helper()

	buildkitName := domain.RunnerBuildkitServiceName(runnerContainer)
	volumeName := domain.RunnerBuildkitStateVolumeName(runnerContainer)

	networks := innerDockerLines(t, env, dindName, "network", "ls", "--format", "{{.Name}}")
	require.Contains(t, networks, buildkitName)

	running := innerDockerLines(t, env, dindName, "ps", "--format", "{{.Names}}")
	require.Contains(t, running, buildkitName)

	aliasFormat := `{{range $name, $net := .NetworkSettings.Networks}}{{$name}}={{range $net.Aliases}}{{.}} {{end}}{{end}}`
	aliases, exitCode := execInDind(t, env, dindName, "inspect", "--format", aliasFormat, buildkitName)
	require.Zero(t, exitCode, aliases)
	require.Contains(t, aliases, buildkitName+"=")
	require.Contains(t, aliases, domain.RunnerBuildkitAlias)

	volumes := innerDockerLines(t, env, dindName, "volume", "ls", "--format", "{{.Name}}")
	require.Contains(t, volumes, volumeName)

	requireRunnerNetworkMode(t, env, runnerContainer, buildkitName)
}

func requireBuildkitReachable(t *testing.T, env *TestEnvironment, dindName, runnerContainer string) {
	t.Helper()

	probeArgs := newBuildctlProbeArgs(domain.RunnerBuildkitServiceName(runnerContainer))

	require.Eventually(t, func() bool {
		_, exitCode := execInDind(t, env, dindName, probeArgs...)

		return exitCode == 0
	}, buildkitProbeWindow, buildkitProbeTick, "buildctl never reached tcp://buildkit:1234 from the buildkit network")
}

func newBuildctlProbeArgs(networkName string) []string {
	addr := "tcp://" + domain.RunnerBuildkitAlias + ":1234"

	return []string{
		"run", "--rm", "--network", networkName,
		"--entrypoint", "buildctl", buildkitImage,
		"--addr", addr, "debug", "workers",
	}
}

func requireBuildkitGoneFromDind(t *testing.T, env *TestEnvironment, dindName, runnerContainer string) {
	t.Helper()

	buildkitName := domain.RunnerBuildkitServiceName(runnerContainer)
	volumeName := domain.RunnerBuildkitStateVolumeName(runnerContainer)

	containers := innerDockerLines(t, env, dindName, "ps", "-a", "--format", "{{.Names}}")
	require.NotContains(t, containers, buildkitName)

	networks := innerDockerLines(t, env, dindName, "network", "ls", "--format", "{{.Name}}")
	require.NotContains(t, networks, buildkitName)

	volumes := innerDockerLines(t, env, dindName, "volume", "ls", "--format", "{{.Name}}")
	require.NotContains(t, volumes, volumeName)

	requireRunnerNetworkMode(t, env, runnerContainer, "")
}

func Test_RunnerBuildkit_Lifecycle_RunsInsideDind(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runRunnerBuildkitLifecycle)
}

func runRunnerBuildkitLifecycle(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := createRunnerOnDind(t, env)

	logDindOnFailure(t, env, fixture.dindName, fixture.runnerContainer)
	preloadBuildkitImage(t, env, fixture.dindName)

	t.Run("enable", func(t *testing.T) {
		setRunnerBuildkitAndAwait(t, env, fixture.runnerContainer, true)

		requireBuildkitInsideDind(t, env, fixture.dindName, fixture.runnerContainer)
	})

	t.Run("buildkitd answers inside the buildkit network", func(t *testing.T) {
		requireBuildkitReachable(t, env, fixture.dindName, fixture.runnerContainer)
	})

	t.Run("enable again is idempotent", func(t *testing.T) {
		setRunnerBuildkitAndAwait(t, env, fixture.runnerContainer, true)

		requireBuildkitInsideDind(t, env, fixture.dindName, fixture.runnerContainer)
	})

	t.Run("disable", func(t *testing.T) {
		setRunnerBuildkitAndAwait(t, env, fixture.runnerContainer, false)

		requireBuildkitGoneFromDind(t, env, fixture.dindName, fixture.runnerContainer)
	})

	t.Run("disable again is idempotent", func(t *testing.T) {
		setRunnerBuildkitAndAwait(t, env, fixture.runnerContainer, false)

		requireBuildkitGoneFromDind(t, env, fixture.dindName, fixture.runnerContainer)
	})
}

func Test_RunnerBuildkit_NestedRuntime_TalksToInnerDaemon(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)
	dindName := dindE2eName(t)

	registerDindCleanup(t, env, dindName)
	createDindAndAwait(t, env, newCreateDindRequest(dindName, toolbox.ToPtr(false)))

	nested, closer, err := nodeResolver(env).NestedRuntime(t.Context(), "", dindName)
	require.NoError(t, err)

	t.Cleanup(func() { _ = closer.Close() })

	require.Eventually(t, func() bool {
		_, listErr := nested.ListNetworks(t.Context(), true)

		return listErr == nil
	}, buildkitProbeWindow, buildkitProbeTick, "inner daemon never answered through the tunnel")

	err = nested.CreateNetwork(t.Context(), tunnelProbeNetwork)
	require.NoError(t, err)

	networks := innerDockerLines(t, env, dindName, "network", "ls", "--format", "{{.Name}}")
	require.Contains(t, networks, tunnelProbeNetwork)

	requireNetworkAbsentOnNode(t, env, tunnelProbeNetwork)

	inspected, isFound, err := nested.Inspect(t.Context(), dindName)
	require.NoError(t, err)
	require.False(t, isFound, "the dind container itself lives on the node, not inside the dind")
	require.Nil(t, inspected.ContainerJSONBase)
}

func requireNetworkAbsentOnNode(t *testing.T, env *TestEnvironment, name string) {
	t.Helper()

	nodeClient := env.Custom.NodeClients.Docker().Client()

	_, err := nodeClient.NetworkInspect(context.Background(), name, network.InspectOptions{})
	require.True(t, client.IsErrNotFound(err), "network %q must not exist on the node, got %v", name, err)
}

func Test_RunnerBuildkit_RunnerWithoutDind_IsRejected(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runRunnerBuildkitWithoutDind)
}

func runRunnerBuildkitWithoutDind(t *testing.T, env *TestEnvironment, _ Plane) {
	runnerContainer := createRunnerWithoutDind(t, env)

	_, err := env.Custom.Services.Runners().SetRunnerBuildkit(t.Context(), runnerContainer, true)
	require.ErrorIs(t, err, user_errors.ErrRunnerBuildkitRequiresDind)
}

func createRunnerWithoutDind(t *testing.T, env *TestEnvironment) string {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()
	runnersClient := velez_api.NewRunnersAPIClient(env.grpcConn)

	suffix := dindE2eName(t)
	runnerName := strings.Replace(suffix, dindE2eNamePrefix, "e2e-runner-", 1)
	runnerContainer := labels.GitlabRunnerNamePrefix + runnerName
	socketAddress := "tcp://" + suffix + "-external:2375"

	t.Cleanup(func() {
		_, _ = runnersClient.DropRunner(context.Background(), newDropRunnerRequest(runnerContainer))

		removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}
		_ = dockerClient.ContainerRemove(context.Background(), runnerContainer, removeOpts)
	})

	baseUrl := startGitlabStubOnBridge(t, env, gitlabStubContainerName(suffix))

	req := newGitlabSocketCreateRunnerRequest(runnerName, baseUrl, socketAddress)

	created, err := runnersClient.CreateRunner(t.Context(), req)
	require.NoError(t, err)

	task := awaitJobTask(t, env, created.GetEntityId(), created.GetAction())
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "create runner task error: %s", task.Error.String)

	return runnerContainer
}
