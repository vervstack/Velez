//go:build e2e_full

package e2e

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/toolbox"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	dindE2eNamePrefix     = "e2e-dind-"
	sysboxRuntimeName     = "sysbox-runc"
	dindTaskTimeout       = 8 * time.Minute
	dindReadyTimeout      = 2 * time.Minute
	runnerRedeployTimeout = 3 * time.Minute
	dindReadyTick         = 2 * time.Second
	dindDockerSockPath    = "/var/run/docker.sock"

	planeClusterSuffix = ""

	sysboxSmokeNamePrefix  = "velez-sysbox-smoke-"
	sysboxSmokeCleanupWait = 30 * time.Second
	sysboxSmokeCleanupTick = time.Second

	runnerDindTarget       = "acme/app"
	runnerDindAccessToken  = "glpat-e2e"
	gitlabStubNamePart     = "gitlab-stub"
	gitlabStubPort         = "80/tcp"
	gitlabStubHealthPath   = "/healthz"
	gitlabStubHost         = "127.0.0.1"
	gitlabStubRunnerId     = 456
	gitlabStubRunnerToken  = "glrt-e2etoken"
	gitlabStubRotatedToken = "glrt-rotated"
	gitlabStubNginxConf    = `log_format stub '$request_method $request_uri $http_private_token';
server {
  listen 80;
  access_log /dev/stdout stub;
  default_type application/json;
  location = /healthz { access_log off; return 200; }
  location /api/v4/projects/ { return 200 '{"id":123}'; }
  location /api/v4/groups/ { return 200 '{"id":321}'; }
  location = /api/v4/user/runners { return 201 '{"id":456,"token":"glrt-e2etoken"}'; }
  location = /api/v4/runners { return 204; }
  location = /api/v4/runners/reset_authentication_token {
    return 201 '{"token":"glrt-rotated","token_expires_at":null}';
  }
}`
)

func dindE2eName(t *testing.T) string {
	t.Helper()

	sum := sha1.Sum([]byte(t.Name()))

	return dindE2eNamePrefix + hex.EncodeToString(sum[:4])
}

func newGetSettingsRequest() *velez_api.GetSettings_Request {
	return &velez_api.GetSettings_Request{}
}

func newUpdateSettingsRequest(isSysboxEnabled, isWhitelistIgnored *bool) *velez_api.UpdateSettings_Request {
	return &velez_api.UpdateSettings_Request{
		IsSysboxEnabled:          isSysboxEnabled,
		IsSysboxWhitelistIgnored: isWhitelistIgnored,
	}
}

func newGetSysboxStatusRequest() *velez_api.GetSysboxStatus_Request {
	return &velez_api.GetSysboxStatus_Request{}
}

func newRunSysboxSmokeTestRequest() *velez_api.RunSysboxSmokeTest_Request {
	return &velez_api.RunSysboxSmokeTest_Request{}
}

func newCreateDindRequest(name string, isSysboxEnabled *bool) *velez_api.CreateDind_Request {
	return &velez_api.CreateDind_Request{
		Name:            name,
		IsSysboxEnabled: isSysboxEnabled,
	}
}

func newDropDindRequest(name string) *velez_api.DropDind_Request {
	return &velez_api.DropDind_Request{Name: name}
}

func newListDindsRequest() *velez_api.ListDinds_Request {
	return &velez_api.ListDinds_Request{}
}

func newCreateRunnerRequest(name string, dindName, socketAddress *string) *velez_api.CreateRunner_Request {
	githubConfig := &velez_api.GithubConfig{AccessToken: runnerDindAccessToken}

	return &velez_api.CreateRunner_Request{
		Name:                name,
		Scope:               velez_api.RunnerScope_REPO,
		Target:              runnerDindTarget,
		DindName:            dindName,
		DockerSocketAddress: socketAddress,
		ProviderConfig:      &velez_api.CreateRunner_Request_Github{Github: githubConfig},
	}
}

func newGitlabCreateRunnerRequest(name, dindName, baseUrl string) *velez_api.CreateRunner_Request {
	gitlabConfig := &velez_api.GitlabConfig{
		AccessToken: runnerDindAccessToken,
		BaseUrl:     toolbox.ToPtr(baseUrl),
		Concurrent:  toolbox.ToPtr(int32(1)),
	}

	return &velez_api.CreateRunner_Request{
		Name:           name,
		Scope:          velez_api.RunnerScope_REPO,
		Target:         runnerDindTarget,
		DindName:       toolbox.ToPtr(dindName),
		ProviderConfig: &velez_api.CreateRunner_Request_Gitlab{Gitlab: gitlabConfig},
	}
}

func newDropRunnerRequest(name string) *velez_api.DropRunner_Request {
	return &velez_api.DropRunner_Request{Name: name}
}

func requireSettings(t *testing.T, env *TestEnvironment, isSysboxEnabled, isWhitelistIgnored bool) {
	t.Helper()

	settingsClient := velez_api.NewSettingsAPIClient(env.grpcConn)

	resp, err := settingsClient.GetSettings(t.Context(), newGetSettingsRequest())
	require.NoError(t, err)
	require.Equal(t, isSysboxEnabled, resp.GetSettings().GetIsSysboxEnabled())
	require.Equal(t, isWhitelistIgnored, resp.GetSettings().GetIsSysboxWhitelistIgnored())
}

func updateSettings(t *testing.T, env *TestEnvironment, req *velez_api.UpdateSettings_Request) {
	t.Helper()

	settingsClient := velez_api.NewSettingsAPIClient(env.grpcConn)

	_, err := settingsClient.UpdateSettings(t.Context(), req)
	require.NoError(t, err)
}

func restoreSettingsOnCleanup(t *testing.T, env *TestEnvironment) {
	t.Helper()

	t.Cleanup(func() {
		settingsClient := velez_api.NewSettingsAPIClient(env.grpcConn)
		req := newUpdateSettingsRequest(toolbox.ToPtr(false), toolbox.ToPtr(false))

		_, _ = settingsClient.UpdateSettings(context.Background(), req)
	})
}

// newPlaneEnvironment builds env for plane. A cluster plane additionally
// enables the statefull plugin, so settings, dind_instances and runners are
// served by the Postgres-backed storages instead of the local state/labels.
// Safe from parallel tests: the cluster-pg port is leased per call, and the
// unsuffixed sidecar name is held exclusively until the test ends.
func newPlaneEnvironment(t *testing.T, plane Plane) *TestEnvironment {
	t.Helper()

	if plane.Mode != ModeCluster {
		return plane.NewEnvironment(t)
	}

	env, _ := enableStatefullPgUnderDind(t, plane, planeClusterSuffix)

	return env
}

func runOnEveryPlane(t *testing.T, body func(t *testing.T, env *TestEnvironment, plane Plane)) {
	t.Helper()

	for _, plane := range Planes {
		t.Run(plane.Name(), func(t *testing.T) {
			env := newPlaneEnvironment(t, plane)

			body(t, env, plane)
		})
	}
}

func isSysboxRuntimeRegistered(t *testing.T, env *TestEnvironment) bool {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	isRegistered, err := dockerutils.HasRuntime(t.Context(), dockerClient, sysboxRuntimeName)
	require.NoError(t, err)

	return isRegistered
}

func requireSysboxRuntimeOrSkip(t *testing.T, env *TestEnvironment) {
	t.Helper()

	if !isSysboxRuntimeRegistered(t, env) {
		t.Skipf("%s is not registered in the test Docker daemon (Linux-only runtime)", sysboxRuntimeName)
	}
}

func Test_Settings_RoundTrip(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, func(t *testing.T, env *TestEnvironment, _ Plane) {
		restoreSettingsOnCleanup(t, env)

		requireSettings(t, env, false, false)

		updateSettings(t, env, newUpdateSettingsRequest(nil, toolbox.ToPtr(true)))
		requireSettings(t, env, false, true)

		updateSettings(t, env, newUpdateSettingsRequest(nil, toolbox.ToPtr(false)))
		requireSettings(t, env, false, false)
	})
}

func Test_Settings_PartialUpdate_LeavesOtherFieldUntouched(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, func(t *testing.T, env *TestEnvironment, _ Plane) {
		restoreSettingsOnCleanup(t, env)

		updateSettings(t, env, newUpdateSettingsRequest(nil, toolbox.ToPtr(true)))
		requireSettings(t, env, false, true)

		updateSettings(t, env, newUpdateSettingsRequest(toolbox.ToPtr(false), nil))
		requireSettings(t, env, false, true)

		updateSettings(t, env, newUpdateSettingsRequest(nil, nil))
		requireSettings(t, env, false, true)
	})
}

func Test_Settings_EnableSysbox_WithoutRuntime_FailsAndLeavesSettingOff(t *testing.T) {
	env := Planes[0].NewEnvironment(t)
	restoreSettingsOnCleanup(t, env)

	if isSysboxRuntimeRegistered(t, env) {
		t.Skipf("%s is registered, the unavailable-runtime case does not apply", sysboxRuntimeName)
	}

	settingsClient := velez_api.NewSettingsAPIClient(env.grpcConn)

	_, err := settingsClient.UpdateSettings(t.Context(), newUpdateSettingsRequest(toolbox.ToPtr(true), nil))
	require.Equal(t, codes.FailedPrecondition, status.Code(err))

	requireSettings(t, env, false, false)
}

func Test_Settings_EnableSysbox_WithRuntime_Succeeds(t *testing.T) {
	env := Planes[0].NewEnvironment(t)
	restoreSettingsOnCleanup(t, env)

	requireSysboxRuntimeOrSkip(t, env)

	updateSettings(t, env, newUpdateSettingsRequest(toolbox.ToPtr(true), nil))
	requireSettings(t, env, true, false)
}

func removeDindResources(dockerClient client.APIClient, name string) {
	ctx := context.Background()

	removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}

	_ = dockerClient.ContainerRemove(ctx, name, removeOpts)
	_ = dockerClient.NetworkRemove(ctx, domain.DindNetworkName(name))
	_ = dockerClient.VolumeRemove(ctx, domain.DindDataVolumeName(name), true)
}

func registerDindCleanup(t *testing.T, env *TestEnvironment, name string) {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeDindResources(dockerClient, name)

	t.Cleanup(func() {
		dindClient := velez_api.NewDindAPIClient(env.grpcConn)

		dropResp, err := dindClient.DropDind(context.Background(), newDropDindRequest(name))
		if err != nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), dindTaskTimeout)
		defer cancel()

		for range env.Custom.JobsEngine.Watch(ctx, dropResp.GetEntityId(), dropResp.GetAction()) {
		}

		removeDindResources(dockerClient, name)
	})
}

func awaitJobTask(t *testing.T, env *TestEnvironment, entityId, action string) tasks_queries.VelezTask {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), dindTaskTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	for task := range env.Custom.JobsEngine.Watch(ctx, entityId, action) {
		finalTask = task
	}

	return finalTask
}

func dropDindAndAwait(t *testing.T, env *TestEnvironment, name string) {
	t.Helper()

	dindClient := velez_api.NewDindAPIClient(env.grpcConn)

	dropResp, err := dindClient.DropDind(t.Context(), newDropDindRequest(name))
	require.NoError(t, err)
	require.Equal(t, name, dropResp.GetEntityId())
	require.Equal(t, jobs.DropDindAction, dropResp.GetAction())

	task := awaitJobTask(t, env, dropResp.GetEntityId(), dropResp.GetAction())
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "drop dind task error: %s", task.Error.String)
}

func createDindAndAwait(t *testing.T, env *TestEnvironment, req *velez_api.CreateDind_Request) {
	t.Helper()

	dindClient := velez_api.NewDindAPIClient(env.grpcConn)

	_, err := dindClient.CreateDind(t.Context(), req)
	require.NoError(t, err)

	task := awaitJobTask(t, env, req.GetName(), jobs.CreateDindAction)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "create dind task error: %s", task.Error.String)
}

func findDind(t *testing.T, env *TestEnvironment, name string) *velez_api.DindInfo {
	t.Helper()

	dindClient := velez_api.NewDindAPIClient(env.grpcConn)

	resp, err := dindClient.ListDinds(t.Context(), newListDindsRequest())
	require.NoError(t, err)

	for _, dind := range resp.GetDinds() {
		if dind.GetName() == name {
			return dind
		}
	}

	return nil
}

func requireDindListed(t *testing.T, env *TestEnvironment, name string) *velez_api.DindInfo {
	t.Helper()

	var found *velez_api.DindInfo

	require.Eventually(t, func() bool {
		found = findDind(t, env, name)

		return found != nil
	}, dindReadyTimeout, dindReadyTick, "dind %q never listed", name)

	return found
}

func requireDaemonResponds(t *testing.T, env *TestEnvironment, containerName string) {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()
	runtimes := container_runtime.NewResolver(dockerClient, env.Custom.NodeClients.Docker().Host(), nil, nil)

	runtime, err := runtimes.Runtime(t.Context(), "")
	require.NoError(t, err)

	execOpts := container.ExecOptions{
		Cmd:          []string{"docker", "-H", "tcp://127.0.0.1:2375", "version"},
		AttachStdout: true,
		AttachStderr: true,
	}

	require.Eventually(t, func() bool {
		_, exitCode, execErr := runtime.Exec(t.Context(), containerName, execOpts)

		return execErr == nil && exitCode == 0
	}, dindReadyTimeout, dindReadyTick, "nested docker daemon never answered on 2375")
}

func requireDindContainerGone(t *testing.T, dockerClient client.APIClient, name string) {
	t.Helper()

	_, err := dockerClient.ContainerInspect(t.Context(), name)
	require.True(t, client.IsErrNotFound(err), "container %q must be gone, got %v", name, err)
}

func requireDindNetworkGone(t *testing.T, dockerClient client.APIClient, name string) {
	t.Helper()

	_, err := dockerClient.NetworkInspect(t.Context(), name, network.InspectOptions{})
	require.True(t, client.IsErrNotFound(err), "network %q must be gone, got %v", name, err)
}

func Test_Dind_Privileged_Lifecycle(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runDindPrivilegedLifecycle)
}

func runDindPrivilegedLifecycle(t *testing.T, env *TestEnvironment, _ Plane) {
	dockerClient := env.Custom.NodeClients.Docker().Client()
	name := dindE2eName(t)
	netName := domain.DindNetworkName(name)

	registerDindCleanup(t, env, name)

	createDindAndAwait(t, env, newCreateDindRequest(name, toolbox.ToPtr(false)))

	info := requireDindListed(t, env, name)
	require.Equal(t, "tcp://"+name+":2375", info.GetAddress())
	require.False(t, info.GetIsSysboxEnabled())
	require.NotNil(t, info.GetCreatedAt())
	require.Contains(t, listedServiceLabels(t, env, name), domain.LabelServiceDind)

	inspected, err := dockerClient.ContainerInspect(t.Context(), name)
	require.NoError(t, err)
	require.Equal(t, domain.DindImage, inspected.Config.Image)
	require.True(t, inspected.HostConfig.Privileged)
	require.NotEqual(t, sysboxRuntimeName, inspected.HostConfig.Runtime)
	require.Contains(t, inspected.Config.Env, "DOCKER_TLS_CERTDIR=")

	attached, isAttached := inspected.NetworkSettings.Networks[netName]
	require.True(t, isAttached, "dind must be attached to %q", netName)
	require.Contains(t, attached.Aliases, name)

	requireDaemonResponds(t, env, name)

	dropDindAndAwait(t, env, name)

	require.Nil(t, findDind(t, env, name))
	requireDindContainerGone(t, dockerClient, name)
	requireDindNetworkGone(t, dockerClient, netName)

	requireDindDataVolumeGone(t, dockerClient, name)
}

func requireDindDataVolumeGone(t *testing.T, dockerClient client.APIClient, name string) {
	t.Helper()

	volumeName := domain.DindDataVolumeName(name)

	_, err := dockerClient.VolumeInspect(t.Context(), volumeName)
	require.True(t, client.IsErrNotFound(err), "volume %q must be gone, got %v", volumeName, err)
}

func Test_Dind_Create_RejectsBadRequests(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)
	dindClient := velez_api.NewDindAPIClient(env.grpcConn)

	t.Run("name must be dns safe", func(t *testing.T) {
		t.Parallel()

		_, err := dindClient.CreateDind(t.Context(), newCreateDindRequest("Bad_Name", toolbox.ToPtr(false)))
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("dropping an unknown dind is not found", func(t *testing.T) {
		t.Parallel()

		_, err := dindClient.DropDind(t.Context(), newDropDindRequest("e2e-dind-unknown"))
		require.Equal(t, codes.NotFound, status.Code(err))
	})
}

func Test_Dind_Create_SysboxDefaultWithoutRuntime_FailsPrecondition(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	if isSysboxRuntimeRegistered(t, env) {
		t.Skipf("%s is registered, the unavailable-runtime case does not apply", sysboxRuntimeName)
	}

	dindClient := velez_api.NewDindAPIClient(env.grpcConn)

	_, err := dindClient.CreateDind(t.Context(), newCreateDindRequest(dindE2eName(t), nil))
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func Test_Dind_Sysbox_Lifecycle(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	requireSysboxRuntimeOrSkip(t, env)

	dockerClient := env.Custom.NodeClients.Docker().Client()
	name := dindE2eName(t)

	registerDindCleanup(t, env, name)

	createDindAndAwait(t, env, newCreateDindRequest(name, nil))

	info := requireDindListed(t, env, name)
	require.True(t, info.GetIsSysboxEnabled())

	inspected, err := dockerClient.ContainerInspect(t.Context(), name)
	require.NoError(t, err)
	require.Equal(t, sysboxRuntimeName, inspected.HostConfig.Runtime)
	require.False(t, inspected.HostConfig.Privileged)

	requireDaemonResponds(t, env, name)
}

func Test_Settings_SysboxEnabled_PlainContainerGetsSysboxRuntime(t *testing.T) {
	env := Planes[0].NewEnvironment(t)
	restoreSettingsOnCleanup(t, env)

	requireSysboxRuntimeOrSkip(t, env)

	updateSettings(t, env, newUpdateSettingsRequest(toolbox.ToPtr(true), nil))

	dockerClient := env.Custom.NodeClients.Docker().Client()
	name := GetServiceName(t)

	smerdReq := &velez_api.CreateSmerd_Request{Name: name, ImageName: NginxAlpineImage}
	smerd := env.CreateSmerd(t, smerdReq)
	require.NotEmpty(t, smerd.GetUuid())

	inspected, err := dockerClient.ContainerInspect(t.Context(), smerd.GetUuid())
	require.NoError(t, err)
	require.Equal(t, sysboxRuntimeName, inspected.HostConfig.Runtime)
}

func Test_Runner_Create_RejectsBadDockerSource(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runRunnerRejectsBadDockerSource)
}

func runRunnerRejectsBadDockerSource(t *testing.T, env *TestEnvironment, _ Plane) {
	runnersClient := velez_api.NewRunnersAPIClient(env.grpcConn)

	t.Run("neither dind_name nor docker_socket_address", func(t *testing.T) {
		t.Parallel()

		req := newCreateRunnerRequest("e2e-runner-neither", nil, nil)

		_, err := runnersClient.CreateRunner(t.Context(), req)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("both dind_name and docker_socket_address", func(t *testing.T) {
		t.Parallel()

		req := newCreateRunnerRequest("e2e-runner-both", toolbox.ToPtr("some-dind"), toolbox.ToPtr("tcp://dind:2375"))

		_, err := runnersClient.CreateRunner(t.Context(), req)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("unknown dind_name", func(t *testing.T) {
		t.Parallel()

		req := newCreateRunnerRequest("e2e-runner-unknown", toolbox.ToPtr("e2e-dind-unknown"), nil)

		_, err := runnersClient.CreateRunner(t.Context(), req)
		require.Equal(t, codes.NotFound, status.Code(err))
	})
}

func logContainerOnFailure(t *testing.T, dockerClient client.APIClient, name string) {
	t.Helper()

	t.Cleanup(func() {
		if !t.Failed() {
			return
		}

		reader, err := dockerClient.ContainerLogs(context.Background(), name, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		})
		if err != nil {
			return
		}

		defer func() { _ = reader.Close() }()

		logs, _ := io.ReadAll(reader)
		t.Logf("logs of %s:\n%s", name, logs)
	})
}

// gitlabStubContainerName must not contain the dind's name: container name
// filters are substring matches, and a stub named after the dind is picked up
// when the deploy watcher resolves the dind's own container.
func gitlabStubContainerName(dindName string) string {
	return strings.Replace(dindName, dindE2eNamePrefix, "e2e-"+gitlabStubNamePart+"-", 1)
}

// startGitlabStub publishes the stub on a pool port of the DinD daemon and
// returns its loopback url: Velez runs in the test process and calls the
// GitLab API from the host, through requireDindLoopbackBridge.
func startGitlabStub(t *testing.T, env *TestEnvironment, name string) string {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	pullReader, err := dockerClient.ImagePull(t.Context(), NginxAlpineImage, image.PullOptions{})
	require.NoError(t, err)

	_, err = io.Copy(io.Discard, pullReader)
	require.NoError(t, err)

	err = pullReader.Close()
	require.NoError(t, err)

	hostPort := strconv.Itoa(int(reservePoolHostPort(t, env)))
	requireDindLoopbackBridge(t)

	script := "cat > /etc/nginx/conf.d/default.conf <<'EOF'\n" + gitlabStubNginxConf + "\nEOF\nexec nginx -g 'daemon off;'"

	port := nat.Port(gitlabStubPort)

	cfg := &container.Config{
		Image:        NginxAlpineImage,
		Entrypoint:   []string{"sh", "-c", script},
		Labels:       map[string]string{testCaseNameLabel: t.Name()},
		ExposedPorts: nat.PortSet{port: struct{}{}},
	}

	binding := nat.PortBinding{HostIP: "0.0.0.0", HostPort: hostPort}
	hostCfg := &container.HostConfig{PortBindings: nat.PortMap{port: []nat.PortBinding{binding}}}

	t.Cleanup(func() {
		removeOpts := container.RemoveOptions{Force: true}
		_ = dockerClient.ContainerRemove(context.Background(), name, removeOpts)
	})

	_, err = dockerClient.ContainerCreate(t.Context(), cfg, hostCfg, nil, nil, name)
	require.NoError(t, err)

	logContainerOnFailure(t, dockerClient, name)

	err = dockerClient.ContainerStart(t.Context(), name, container.StartOptions{})
	require.NoError(t, err)

	baseUrl := "http://" + net.JoinHostPort(gitlabStubHost, hostPort)

	require.Eventually(t, func() bool {
		resp, getErr := http.Get(baseUrl + gitlabStubHealthPath)
		if getErr != nil {
			return false
		}

		_ = resp.Body.Close()

		return resp.StatusCode == http.StatusOK
	}, dindReadyTimeout, time.Second/4, "gitlab stub never became ready")

	return baseUrl
}

// gitlabStubRequests returns the stub's access log, one "METHOD URI PRIVATE-TOKEN" line per request.
func gitlabStubRequests(t *testing.T, env *TestEnvironment, name string) []string {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	reader, err := dockerClient.ContainerLogs(t.Context(), name, container.LogsOptions{ShowStdout: true})
	require.NoError(t, err)

	defer func() { _ = reader.Close() }()

	var stdout bytes.Buffer

	_, err = stdcopy.StdCopy(&stdout, io.Discard, reader)
	require.NoError(t, err)

	trimmed := strings.TrimSpace(stdout.String())
	if trimmed == "" {
		return nil
	}

	return strings.Split(trimmed, "\n")
}

func newRedeployRunnerRequest(name string) *velez_api.RedeployRunner_Request {
	return &velez_api.RedeployRunner_Request{Name: name}
}

func requireNoDockerSocketMount(t *testing.T, inspected container.InspectResponse) {
	t.Helper()

	for _, mount := range inspected.Mounts {
		require.NotContains(t, mount.Source, "docker.sock")
		require.NotEqual(t, dindDockerSockPath, mount.Destination)
	}

	for _, bind := range inspected.HostConfig.Binds {
		require.NotContains(t, bind, "docker.sock")
	}
}

func requireRunnerOnDind(t *testing.T, inspected container.InspectResponse, dindName string) {
	t.Helper()

	require.Contains(t, inspected.Config.Env, "DOCKER_HOST="+domain.DindAddress(dindName))
	require.Contains(t, inspected.NetworkSettings.Networks, domain.DindNetworkName(dindName))
	requireNoDockerSocketMount(t, inspected)
}

func redeployRunnerAndAwaitNewContainer(
	t *testing.T, env *TestEnvironment, runnerContainer, previousContainerId string,
) container.InspectResponse {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()
	runnersClient := velez_api.NewRunnersAPIClient(env.grpcConn)

	_, err := runnersClient.RedeployRunner(t.Context(), newRedeployRunnerRequest(runnerContainer))
	require.NoError(t, err)

	var redeployed container.InspectResponse

	require.Eventually(t, func() bool {
		current, inspectErr := dockerClient.ContainerInspect(t.Context(), runnerContainer)
		if inspectErr != nil || current.ContainerJSONBase == nil || current.State == nil {
			return false
		}

		redeployed = current

		return current.ID != previousContainerId && current.State.Running
	}, runnerRedeployTimeout, dindReadyTick, "runner container was never redeployed")

	return redeployed
}

func Test_Runner_OnDind_WiresDockerHostAndBlocksDindDrop(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runRunnerOnDind)
}

type dindRunnerFixture struct {
	dindName        string
	runnerContainer string
	stubName        string
	inspected       container.InspectResponse
}

func createRunnerOnDind(t *testing.T, env *TestEnvironment) dindRunnerFixture {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()
	runnersClient := velez_api.NewRunnersAPIClient(env.grpcConn)

	dindName := dindE2eName(t)
	runnerName := strings.Replace(dindName, dindE2eNamePrefix, "e2e-runner-", 1)
	runnerContainer := labels.GitlabRunnerNamePrefix + runnerName

	registerDindCleanup(t, env, dindName)

	t.Cleanup(func() {
		_, _ = runnersClient.DropRunner(context.Background(), newDropRunnerRequest(runnerContainer))

		removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}
		_ = dockerClient.ContainerRemove(context.Background(), runnerContainer, removeOpts)
	})

	createDindAndAwait(t, env, newCreateDindRequest(dindName, toolbox.ToPtr(false)))
	stubName := gitlabStubContainerName(dindName)
	baseUrl := startGitlabStub(t, env, stubName)

	logContainerOnFailure(t, dockerClient, runnerContainer)

	created, err := runnersClient.CreateRunner(t.Context(), newGitlabCreateRunnerRequest(runnerName, dindName, baseUrl))
	require.NoError(t, err)

	task := awaitJobTask(t, env, created.GetEntityId(), created.GetAction())
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "create runner task error: %s", task.Error.String)

	inspected, err := dockerClient.ContainerInspect(t.Context(), runnerContainer)
	require.NoError(t, err)

	return dindRunnerFixture{
		dindName:        dindName,
		runnerContainer: runnerContainer,
		stubName:        stubName,
		inspected:       inspected,
	}
}

func runRunnerOnDind(t *testing.T, env *TestEnvironment, _ Plane) {
	dindClient := velez_api.NewDindAPIClient(env.grpcConn)
	runnersClient := velez_api.NewRunnersAPIClient(env.grpcConn)

	fixture := createRunnerOnDind(t, env)

	requireRunnerOnDind(t, fixture.inspected, fixture.dindName)

	_, err := dindClient.DropDind(t.Context(), newDropDindRequest(fixture.dindName))
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
	requireDindListed(t, env, fixture.dindName)

	dropped, err := runnersClient.DropRunner(t.Context(), newDropRunnerRequest(fixture.runnerContainer))
	require.NoError(t, err)

	dropTask := awaitJobTask(t, env, dropped.GetEntityId(), dropped.GetAction())
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, dropTask.Status, "drop runner task error: %s", dropTask.Error.String)

	dropDindAndAwait(t, env, fixture.dindName)
	require.Nil(t, findDind(t, env, fixture.dindName))
}

func Test_Runner_RedeployOnDind_KeepsDockerHostAndDindNetwork(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runRunnerRedeployOnDind)
}

func runRunnerRedeployOnDind(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := createRunnerOnDind(t, env)

	redeployed := redeployRunnerAndAwaitNewContainer(t, env, fixture.runnerContainer, fixture.inspected.ID)

	requireRunnerOnDind(t, redeployed, fixture.dindName)
}

func newGitlabSocketCreateRunnerRequest(name, baseUrl, socketAddress string) *velez_api.CreateRunner_Request {
	gitlabConfig := &velez_api.GitlabConfig{
		AccessToken: runnerDindAccessToken,
		BaseUrl:     toolbox.ToPtr(baseUrl),
		Concurrent:  toolbox.ToPtr(int32(1)),
	}

	return &velez_api.CreateRunner_Request{
		Name:                name,
		Scope:               velez_api.RunnerScope_REPO,
		Target:              runnerDindTarget,
		DockerSocketAddress: toolbox.ToPtr(socketAddress),
		ProviderConfig:      &velez_api.CreateRunner_Request_Gitlab{Gitlab: gitlabConfig},
	}
}

// The single-node runners storage rebuilds a live row from container labels,
// which do not carry docker_socket_address, so only the Postgres row can be
// checked for it.
func requireRunnerRowWithoutDindLink(
	t *testing.T, env *TestEnvironment, plane Plane, runnerContainer, socketAddress string,
) {
	t.Helper()

	dataStorage := env.Custom.Services.StorageContainer()

	svc, err := dataStorage.Services().GetByName(t.Context(), runnerContainer)
	require.NoError(t, err)

	runner, err := dataStorage.Runners().GetRunnerByServiceID(t.Context(), svc.ID)
	require.NoError(t, err)
	require.Zero(t, runner.DindServiceId)

	if plane.Mode == ModeCluster {
		require.Equal(t, socketAddress, runner.DockerSocketAddress)
	}
}

func Test_Runner_WithTcpDockerSocketAddress_KeepsDockerHostWithoutSocketMountOrDindLink(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runRunnerWithTcpSocketAddress)
}

func runRunnerWithTcpSocketAddress(t *testing.T, env *TestEnvironment, plane Plane) {
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

	baseUrl := startGitlabStub(t, env, gitlabStubContainerName(suffix))

	logContainerOnFailure(t, dockerClient, runnerContainer)

	req := newGitlabSocketCreateRunnerRequest(runnerName, baseUrl, socketAddress)

	created, err := runnersClient.CreateRunner(t.Context(), req)
	require.NoError(t, err)

	task := awaitJobTask(t, env, created.GetEntityId(), created.GetAction())
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "create runner task error: %s", task.Error.String)

	inspected, err := dockerClient.ContainerInspect(t.Context(), runnerContainer)
	require.NoError(t, err)

	require.Contains(t, inspected.Config.Env, "DOCKER_HOST="+socketAddress)
	requireNoDockerSocketMount(t, inspected)
	requireRunnerRowWithoutDindLink(t, env, plane, runnerContainer, socketAddress)
}

func requireNoSysboxSmokeContainers(t *testing.T, env *TestEnvironment) {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()
	listOpts := container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", sysboxSmokeNamePrefix)),
	}

	require.Eventually(t, func() bool {
		list, err := dockerClient.ContainerList(t.Context(), listOpts)

		return err == nil && len(list) == 0
	}, sysboxSmokeCleanupWait, sysboxSmokeCleanupTick, "smoke test container must be removed")
}

func Test_SysboxStatus_ReportsDaemonFacts(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)
	settingsClient := velez_api.NewSettingsAPIClient(env.grpcConn)

	smerdReq := &velez_api.CreateSmerd_Request{Name: GetServiceName(t), ImageName: NginxAlpineImage}
	env.CreateSmerd(t, smerdReq)

	resp, err := settingsClient.GetSysboxStatus(t.Context(), newGetSysboxStatusRequest())
	require.NoError(t, err)

	require.NotEmpty(t, resp.GetOsType())
	require.NotEmpty(t, resp.GetKernelVersion())
	require.NotEmpty(t, resp.GetDockerVersion())
	require.Equal(t, isSysboxRuntimeRegistered(t, env), resp.GetIsRuntimeRegistered())
	require.GreaterOrEqual(t, resp.GetContainersTotal(), int32(1))
	require.GreaterOrEqual(t, resp.GetContainersOnSysbox(), int32(0))
	require.GreaterOrEqual(t, resp.GetContainersTotal(), resp.GetContainersOnSysbox())
}

func Test_SysboxSmokeTest_FailsWithoutRuntime(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	if isSysboxRuntimeRegistered(t, env) {
		t.Skipf("%s is registered, the unavailable-runtime case does not apply", sysboxRuntimeName)
	}

	settingsClient := velez_api.NewSettingsAPIClient(env.grpcConn)

	resp, err := settingsClient.RunSysboxSmokeTest(t.Context(), newRunSysboxSmokeTestRequest())
	require.NoError(t, err)
	require.False(t, resp.GetIsPassed())
	require.NotEmpty(t, resp.GetFailure())

	requireNoSysboxSmokeContainers(t, env)
}

func Test_SysboxSmokeTest_PassesWithRuntime(t *testing.T) {
	env := Planes[0].NewEnvironment(t)

	requireSysboxRuntimeOrSkip(t, env)

	settingsClient := velez_api.NewSettingsAPIClient(env.grpcConn)

	resp, err := settingsClient.RunSysboxSmokeTest(t.Context(), newRunSysboxSmokeTestRequest())
	require.NoError(t, err)
	require.True(t, resp.GetIsPassed(), "smoke test failure: %s", resp.GetFailure())

	requireNoSysboxSmokeContainers(t, env)
}
