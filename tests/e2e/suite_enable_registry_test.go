//go:build e2e_full

package e2e

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
)

// registryPasswordSecretRef mirrors enable_registry.go's own (unexported)
// registryPasswordSecretRef helper - the ref is a fixed, documented triple
// (docs/features/pgaas_and_registry_plugin.md section 1), not worth
// exporting just for this test to import.
func registryPasswordSecretRef() domain.SecretRef {
	return domain.SecretRef{Scope: "plugin", Owner: "registry", Key: "password"}
}

// EnableRegistrySuite exercises the registry plugin end to end: enable it in
// single-node mode (the mode the "doesn't show up on control plane page"
// bug report was about), wait for the deploy watcher to actually create the
// container (see enableRegistryUnderDind's doc comment), then push a real
// image into it and pull it back out, authenticating with the credentials
// the enable flow generated.
//
// Not t.Parallel(): jobs.RegistryServiceName is a fixed, unsuffixed name (a
// node only ever runs one registry - see enable_registry.go), so only one
// instance of this suite can run against a given Docker host at a time. The
// happy-path test cleans up its own fixed-name container/volume
// unconditionally via enableRegistryUnderDind.
type EnableRegistrySuite struct {
	suite.Suite
}

func (s *EnableRegistrySuite) Test_EnableRegistry_HappyPath() {
	t := s.T()
	ctx := t.Context()

	env := enableRegistryUnderDind(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	inspect, err := dockerClient.ContainerInspect(ctx, jobs.RegistryServiceName)
	require.NoError(t, err)
	require.NotNil(t, inspect.State)
	require.True(t, inspect.State.Running)

	// registryContainerName's mapping into pluginContainerNames
	// (internal/storage/local_storage/docker_impl.go) is the fix for the
	// bug where the registry plugin never showed up on the control-plane
	// page in single-node mode - assert the full ListPlugins chain reports
	// it running now.
	listPluginsResp, err := env.Custom.ControlPlaneApiImpl.ListPlugins(ctx, &velez_api.ListPlugins_Request{})
	require.NoError(t, err)

	var registryPlugin *velez_api.Plugin

	for _, plugin := range listPluginsResp.GetPlugins() {
		if plugin.GetType() == velez_api.VervPluginType_registry {
			registryPlugin = plugin

			break
		}
	}

	require.NotNil(t, registryPlugin, "expected registry plugin in ListPlugins response")
	require.Equal(t, velez_api.VervPlugin_running, registryPlugin.GetState())

	// Goes through the real RPC (env.Custom.ControlPlaneApiImpl.ListRegistries)
	// rather than env.Custom.Services.StorageContainer().Registries(), which
	// mirrors suite_enable_statefull_test.go's ListDeployments precedent: the
	// registries row register_registry writes goes through
	// ClusterClients.StateManager() (see enable_registry.go's
	// enableRegistryHandler.storageContainer doc comment and services.go's
	// postgresService one - the same split, deliberate, not yet converged in
	// single-node mode), which is also what VervServicesService.Registries()
	// (and so this RPC) reads from - StorageContainer() is a different,
	// disconnected instance in single-node mode and would see nothing here.
	listRegistriesResp, err := env.Custom.ControlPlaneApiImpl.ListRegistries(ctx, &velez_api.ListRegistries_Request{})
	require.NoError(t, err)

	var registryRow *velez_api.Registry

	for _, reg := range listRegistriesResp.GetRegistries() {
		if reg.GetName() == jobs.RegistryServiceName {
			registryRow = reg

			break
		}
	}

	require.NotNil(t, registryRow, "expected a registries row named %q", jobs.RegistryServiceName)

	password, err := env.Custom.Services.Secrets().Get(ctx, registryPasswordSecretRef())
	require.NoError(t, err)
	require.NotEmpty(t, password)

	authConfig := registry.AuthConfig{
		Username: registryRow.GetUsername(),
		Password: password,
	}

	encodedAuth, err := registry.EncodeAuthConfig(authConfig)
	require.NoError(t, err)

	// registryRow.Url is registryUrl()'s own output (enable_registry.go) -
	// "http://localhost:<port>" for a non-containerized Velez, where <port>
	// is whatever deployRegistryJob.resolvePort actually got handed by
	// PortManager (see enableRegistryUnderDind's doc comment on why this
	// can't be a fixed, pre-chosen port). The push/pull below are performed
	// by dockerd itself (inside the DinD daemon), not by this test process -
	// the docker client only submits the API call, so "localhost:<port>" is
	// the address the daemon's own network namespace sees the container's
	// published port on, not a bootstrap-host address this process would
	// need dindHostAddr to translate.
	registryUrl, err := url.Parse(registryRow.GetUrl())
	require.NoError(t, err)

	registryHost := registryUrl.Host
	targetRef := registryHost + "/e2e-hello-world:pushed"

	// PullImage no-ops if the image is already present - other suites
	// sharing this DinD daemon (see main_test.go) may have already pulled
	// it, but this suite must not depend on running after one that has.
	_, err = dockerutils.PullImage(ctx, dockerClient, HelloWorldAppImage, false)
	require.NoError(t, err)

	err = dockerClient.ImageTag(ctx, HelloWorldAppImage, targetRef)
	require.NoError(t, err)

	pushOpts := image.PushOptions{RegistryAuth: encodedAuth}

	pushRdr, err := dockerClient.ImagePush(ctx, targetRef, pushOpts)
	require.NoError(t, err)

	requireNoStreamErrorf(t, pushRdr, "pushing %s", targetRef)

	removeOpts := image.RemoveOptions{Force: true}

	_, err = dockerClient.ImageRemove(ctx, targetRef, removeOpts)
	require.NoError(t, err)

	pullOpts := image.PullOptions{RegistryAuth: encodedAuth}

	pullRdr, err := dockerClient.ImagePull(ctx, targetRef, pullOpts)
	require.NoError(t, err)

	requireNoStreamErrorf(t, pullRdr, "pulling %s back from the registry", targetRef)

	_, err = dockerClient.ImageInspect(ctx, targetRef)
	require.NoError(t, err, "expected the pulled-back image to be present locally under %s", targetRef)
}

func Test_EnableRegistry(t *testing.T) {
	suite.Run(t, new(EnableRegistrySuite))
}

// dockerStreamMessage captures only the error shape of a Docker API
// push/pull progress message - the small subset of
// github.com/docker/docker/pkg/jsonmessage.JSONMessage this test needs,
// declared locally rather than importing that package, which would pull in
// its terminal-rendering dependencies for one field.
type dockerStreamMessage struct {
	Error       string `json:"error"`
	ErrorDetail struct {
		Message string `json:"message"`
	} `json:"errorDetail"`
}

// requireNoStreamErrorf drains a Docker API push/pull response body, which is
// a stream of newline-delimited JSON progress messages, and fails the test
// if any message carries an embedded error - ImagePush/ImagePull only
// return a Go error for transport-level failures (e.g. connection refused);
// an auth or registry-side failure surfaces solely as one of these
// in-stream messages, same as the `docker push`/`docker pull` CLI.
func requireNoStreamErrorf(t *testing.T, rdr io.ReadCloser, format string, args ...any) {
	t.Helper()

	defer func() {
		closeErr := rdr.Close()
		require.NoError(t, closeErr)
	}()

	scanner := bufio.NewScanner(rdr)

	for scanner.Scan() {
		var msg dockerStreamMessage

		err := json.Unmarshal(scanner.Bytes(), &msg)
		if err != nil {
			continue
		}

		errMsg := msg.Error
		if errMsg == "" {
			errMsg = msg.ErrorDetail.Message
		}

		if errMsg != "" {
			require.Failf(t, "docker stream reported an error", "%s: %s", fmt.Sprintf(format, args...), errMsg)
		}
	}

	require.NoError(t, scanner.Err())
}
