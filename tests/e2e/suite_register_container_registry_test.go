//go:build e2e_full

package e2e

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	registryRegImage         = "registry:2"
	registryRegUser          = "e2e_reg_user"
	registryRegPassword      = "e2e-reg-secret-pw"
	registryRegWrongPassword = "e2e-reg-wrong-pw"
	registryRegDefaultPort   = "5000"

	registryRegContainerPort = nat.Port("5000/tcp")
	registryRegAuthDir       = "auth"
	registryRegHtpasswdFile  = "auth/htpasswd"
	registryRegHtpasswdPath  = "/auth/htpasswd"
	registryRegAuthDirMode   = 0o755
	registryRegHtpasswdMode  = 0o644
	registryRegReadyTimeout  = 90 * time.Second
	registryRegReadyPoll     = time.Second

	registryRegPortSingleFailures = dindPortBandEnd - 1
	registryRegPortSingle         = dindPortBandEnd
	registryRegPortClusterFail    = dindPortBandStart + 9
	registryRegPortCluster        = dindPortBandStart + 10
)

func registryRegNames(prefix string) (containerName, serviceName string) {
	return "e2e_regreg_" + prefix + "_c", "e2e_regreg_" + prefix + "_svc"
}

func registryRegAuthEnv() []string {
	return []string{
		"REGISTRY_AUTH=htpasswd",
		"REGISTRY_AUTH_HTPASSWD_REALM=e2e",
		"REGISTRY_AUTH_HTPASSWD_PATH=" + registryRegHtpasswdPath,
	}
}

func registryRegHtpasswdTar(t *testing.T) io.Reader {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(registryRegPassword), bcrypt.MinCost)
	require.NoError(t, err)

	line := []byte(registryRegUser + ":" + string(hash) + "\n")

	buf := &bytes.Buffer{}
	writer := tar.NewWriter(buf)

	dirHeader := &tar.Header{Name: registryRegAuthDir + "/", Mode: registryRegAuthDirMode, Typeflag: tar.TypeDir}

	err = writer.WriteHeader(dirHeader)
	require.NoError(t, err)

	fileHeader := &tar.Header{Name: registryRegHtpasswdFile, Mode: registryRegHtpasswdMode, Size: int64(len(line))}

	err = writer.WriteHeader(fileHeader)
	require.NoError(t, err)

	_, err = writer.Write(line)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	return bytes.NewReader(buf.Bytes())
}

func pullRegistryImage(t *testing.T, dockerClient client.APIClient) {
	t.Helper()

	pullReader, err := dockerClient.ImagePull(t.Context(), registryRegImage, image.PullOptions{})
	require.NoError(t, err)

	_, err = io.Copy(io.Discard, pullReader)
	require.NoError(t, err)

	err = pullReader.Close()
	require.NoError(t, err)
}

// startForeignRegistry runs a real registry:2 Velez did not create. With
// isAuth it carries htpasswd auth for registryRegUser; a non-zero port
// publishes 5000 there and forwards it to the test process.
func startForeignRegistry(
	t *testing.T, env *TestEnvironment, name string, port int, isAuth bool,
) foreignContainer {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	pullRegistryImage(t, dockerClient)

	cfg := &container.Config{
		Image:  registryRegImage,
		Labels: map[string]string{testCaseNameLabel: t.Name()},
	}

	hostCfg := &container.HostConfig{}

	if isAuth {
		cfg.Env = registryRegAuthEnv()
	}

	if port != 0 {
		binding := nat.PortBinding{HostIP: pgRegAllInterfaces, HostPort: strconv.Itoa(port)}
		hostCfg.PortBindings = nat.PortMap{registryRegContainerPort: []nat.PortBinding{binding}}
	}

	t.Cleanup(func() {
		removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}

		_ = dockerClient.ContainerRemove(context.Background(), name, removeOpts)
	})

	created, err := dockerClient.ContainerCreate(t.Context(), cfg, hostCfg, nil, nil, name)
	require.NoError(t, err)

	if isAuth {
		copyOpts := container.CopyToContainerOptions{}

		err = dockerClient.CopyToContainer(t.Context(), created.ID, "/", registryRegHtpasswdTar(t), copyOpts)
		require.NoError(t, err)
	}

	err = dockerClient.ContainerStart(t.Context(), created.ID, container.StartOptions{})
	require.NoError(t, err)

	if port != 0 {
		forwardDindPort(t, port)
		awaitRegistryReady(t, port)
	}

	return foreignContainer{id: created.ID, name: name}
}

func awaitRegistryReady(t *testing.T, port int) {
	t.Helper()

	target := "http://" + net.JoinHostPort(pgRegLoopback, strconv.Itoa(port)) + "/v2/"

	require.Eventually(t, func() bool {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, target, http.NoBody)
		if err != nil {
			return false
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false
		}

		_ = resp.Body.Close()

		return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized
	}, registryRegReadyTimeout, registryRegReadyPoll)
}

func newRegistryRegisterRequest(
	containerId, serviceName string, username, password *string,
) *velez_api.RegisterContainer_Request {
	req := newRegisterContainerRequest(containerId, serviceName)

	pattern := &velez_api.RegisterContainer_Request_RegistryPattern{Username: username, Password: password}

	req.Pattern = &velez_api.RegisterContainer_Request_Registry{Registry: pattern}

	return req
}

func registryRegSecretRef(serviceName string) domain.SecretRef {
	return domain.SecretRef{Scope: "registryaas", Owner: serviceName, Key: "password"}
}

func requireNoPendingRegistrySecret(t *testing.T, env *TestEnvironment, entityId string) {
	t.Helper()

	ref := jobs.RegisteredRegistryPendingSecretRef(entityId)

	_, err := env.Custom.Services.Secrets().Get(t.Context(), ref)
	require.True(t, rerrors.Is(err, user_errors.ErrSecretNotFound), "pending secret must be gone, got %v", err)
}

func requireNoRegistrySecret(t *testing.T, env *TestEnvironment, serviceName string) {
	t.Helper()

	_, err := env.Custom.Services.Secrets().Get(t.Context(), registryRegSecretRef(serviceName))
	require.True(t, rerrors.Is(err, user_errors.ErrSecretNotFound), "registry secret must not exist, got %v", err)
}

func requireRegistryRegisterNothingChanged(
	t *testing.T,
	env *TestEnvironment,
	foreign foreignContainer,
	serviceName string,
	resp *velez_api.RegisterContainer_Response,
) {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	inspected, err := dockerClient.ContainerInspect(t.Context(), foreign.name)
	require.NoError(t, err)
	require.Equal(t, foreign.id, inspected.ID, "a failed register must not recreate the container")
	require.True(t, inspected.State.Running)
	require.Empty(t, inspected.Config.Labels[labels.VervServiceLabel])
	require.Empty(t, inspected.Config.Labels[labels.RegistryaasInstanceLabel])

	svcResp, err := env.ServiceApiClient().ListServices(t.Context(), newListServicesRequest())
	require.NoError(t, err)

	for _, svc := range svcResp.GetServices() {
		require.NotEqual(t, serviceName, svc.GetName(), "a failed register must leave no service row")
	}

	require.Nil(t, findRegistryInstance(t, env, serviceName), "a failed register must leave no registry instance")

	requireNoRegistrySecret(t, env, serviceName)
	requireNoPendingRegistrySecret(t, env, resp.GetEntityId())
}

func Test_RegisterContainer_Registry_LoginFailuresChangeNothing(t *testing.T) {
	t.Parallel()

	containerName, serviceName := registryRegNames("failures")

	env := Planes[0].NewEnvironment(t)

	foreign := startForeignRegistry(t, env, containerName, registryRegPortSingleFailures, true)

	t.Run("wrong password", func(t *testing.T) {
		req := newRegistryRegisterRequest(
			foreign.id, serviceName, toolbox.ToPtr(registryRegUser), toolbox.ToPtr(registryRegWrongPassword))
		resp := registerContainer(t, env, req)

		task := awaitRegisterTask(t, env, resp)
		require.Equal(t, tasks_queries.VelezTaskStatusFAILED, task.Status)
		require.Contains(t, task.Error.String, user_errors.ErrRegistryLoginFailed.Error())
		requireNoPasswordInPayload(t, task, registryRegWrongPassword)

		requireRegistryRegisterNothingChanged(t, env, foreign, serviceName, resp)
	})

	t.Run("credentials missing", func(t *testing.T) {
		resp := registerContainer(t, env, newRegistryRegisterRequest(foreign.id, serviceName, nil, nil))

		task := awaitRegisterTask(t, env, resp)
		require.Equal(t, tasks_queries.VelezTaskStatusFAILED, task.Status)
		require.Contains(t, task.Error.String, user_errors.ErrRegistryCredentialsRequired.Error())

		requireRegistryRegisterNothingChanged(t, env, foreign, serviceName, resp)
	})
}

func Test_RegisterContainer_Registry_SingleMode(t *testing.T) {
	t.Parallel()

	containerName, serviceName := registryRegNames("single")

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	foreign := startForeignRegistry(t, env, containerName, registryRegPortSingle, true)

	req := newRegistryRegisterRequest(
		foreign.id, serviceName, toolbox.ToPtr(registryRegUser), toolbox.ToPtr(registryRegPassword))
	req.KeepPortMapping = true

	resp := registerContainer(t, env, req)

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "register task error: %s", task.Error.String)
	requireNoPasswordInPayload(t, task, registryRegPassword)

	inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)
	require.NotEqual(t, foreign.id, inspected.ID, "single mode recreates the container")

	containerLabels := inspected.Config.Labels
	require.Equal(t, serviceName, containerLabels[labels.VervServiceLabel])
	require.Equal(t, labelValueTrue, containerLabels[labels.RegistryaasInstanceLabel])
	require.Equal(t, registryRegUser, containerLabels[labels.RegistryaasUsernameLabel])
	require.Equal(t, strconv.Itoa(registryRegPortSingle), containerLabels[labels.RegistryaasPortLabel])

	instance := findRegistryInstance(t, env, serviceName)
	require.NotNil(t, instance, "registered registry must be visible in ListRegistryInstances")
	require.Equal(t, registryRegUser, instance.GetUsername())
	require.EqualValues(t, registryRegPortSingle, instance.GetPort())

	secret, err := env.Custom.Services.Secrets().Get(t.Context(), registryRegSecretRef(serviceName))
	require.NoError(t, err)
	require.Equal(t, registryRegPassword, secret)

	requireNoPendingRegistrySecret(t, env, resp.GetEntityId())
}

func Test_RegisterContainer_Registry_SingleMode_NoAuthNeedsNoCredentials(t *testing.T) {
	t.Parallel()

	containerName, serviceName := registryRegNames("noauth")

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	foreign := startForeignRegistry(t, env, containerName, 0, false)

	resp := registerContainer(t, env, newRegistryRegisterRequest(foreign.id, serviceName, nil, nil))

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "register task error: %s", task.Error.String)

	inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)
	require.NotEqual(t, foreign.id, inspected.ID, "single mode recreates the container")

	containerLabels := inspected.Config.Labels
	require.Equal(t, serviceName, containerLabels[labels.VervServiceLabel])
	require.Equal(t, labelValueTrue, containerLabels[labels.RegistryaasInstanceLabel])
	require.Empty(t, containerLabels[labels.RegistryaasUsernameLabel])
	require.Equal(t, registryRegDefaultPort, containerLabels[labels.RegistryaasPortLabel])

	requireNoRegistrySecret(t, env, serviceName)
	requireNoPendingRegistrySecret(t, env, resp.GetEntityId())
}

func Test_RegisterContainer_Registry_ClusterMode(t *testing.T) {
	env, _ := enableStatefullPgUnderDind(t, Planes[1], registerClusterSuffix+"-registry")

	t.Run("wrong password changes nothing", func(t *testing.T) {
		containerName, serviceName := registryRegNames("cluster_fail")

		foreign := startForeignRegistry(t, env, containerName, registryRegPortClusterFail, true)

		req := newRegistryRegisterRequest(
			foreign.id, serviceName, toolbox.ToPtr(registryRegUser), toolbox.ToPtr(registryRegWrongPassword))
		resp := registerContainer(t, env, req)

		task := awaitRegisterTask(t, env, resp)
		require.Equal(t, tasks_queries.VelezTaskStatusFAILED, task.Status)
		require.Contains(t, task.Error.String, user_errors.ErrRegistryLoginFailed.Error())

		requireRegistryRegisterNothingChanged(t, env, foreign, serviceName, resp)
	})

	t.Run("right password", func(t *testing.T) {
		containerName, serviceName := registryRegNames("cluster_ok")

		foreign := startForeignRegistry(t, env, containerName, registryRegPortCluster, true)

		req := newRegistryRegisterRequest(
			foreign.id, serviceName, toolbox.ToPtr(registryRegUser), toolbox.ToPtr(registryRegPassword))
		resp := registerContainer(t, env, req)

		task := awaitRegisterTask(t, env, resp)
		require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "register task error: %s", task.Error.String)
		requireNoPasswordInPayload(t, task, registryRegPassword)

		requireContainerUntouched(t, env.Custom.NodeClients.Docker().Client(), foreign)
		requireDeploymentRunning(t, env, serviceName)

		instance := findRegistryInstance(t, env, serviceName)
		require.NotNil(t, instance, "registered registry must be visible in ListRegistryInstances")
		require.Equal(t, registryRegUser, instance.GetUsername())
		require.EqualValues(t, registryRegPortCluster, instance.GetPort())

		secret, err := env.Custom.Services.Secrets().Get(t.Context(), registryRegSecretRef(serviceName))
		require.NoError(t, err)
		require.Equal(t, registryRegPassword, secret)

		requireNoPendingRegistrySecret(t, env, resp.GetEntityId())
	})

	t.Run("no auth needs no credentials", func(t *testing.T) {
		containerName, serviceName := registryRegNames("cluster_noauth")

		foreign := startForeignRegistry(t, env, containerName, 0, false)

		resp := registerContainer(t, env, newRegistryRegisterRequest(foreign.id, serviceName, nil, nil))

		task := awaitRegisterTask(t, env, resp)
		require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "register task error: %s", task.Error.String)

		requireContainerUntouched(t, env.Custom.NodeClients.Docker().Client(), foreign)

		instance := findRegistryInstance(t, env, serviceName)
		require.NotNil(t, instance)
		require.Empty(t, instance.GetUsername())

		requireNoRegistrySecret(t, env, serviceName)
	})
}

func Test_RegisterContainer_PendingSecretsDroppedWhenTaskFails(t *testing.T) {
	t.Parallel()

	_, serviceName := registryRegNames("pending")

	env := Planes[0].NewEnvironment(t)

	req := newRegistryRegisterRequest(
		"no-such-container-id", serviceName, toolbox.ToPtr(registryRegUser), toolbox.ToPtr(registryRegPassword))
	resp := registerContainer(t, env, req)

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, tasks_queries.VelezTaskStatusFAILED, task.Status)

	requireNoPendingRegistrySecret(t, env, resp.GetEntityId())
}

func Test_RegisterContainer_Registry_RejectsPartialCredentials(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)
	velezClient := velez_api.NewVelezAPIClient(env.grpcConn)

	t.Run("username without password", func(t *testing.T) {
		req := newRegistryRegisterRequest("some-id", "e2e_regreg_bad", toolbox.ToPtr(registryRegUser), nil)

		_, err := velezClient.RegisterContainer(t.Context(), req)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("password without username", func(t *testing.T) {
		req := newRegistryRegisterRequest("some-id", "e2e_regreg_bad", nil, toolbox.ToPtr(registryRegPassword))

		_, err := velezClient.RegisterContainer(t.Context(), req)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}
