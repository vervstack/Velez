package container_runtime

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	testSmerdName = "mysvc"
	testSuffix    = "stage"
)

// fakeCreateAPI is a minimal hand-written fake for the single client.APIClient
// method the label-based runtime uses. The full interface is embedded as a nil
// value purely so the fake satisfies it - calling anything else panics, same
// as calling a method on a nil interface.
type fakeCreateAPI struct {
	client.APIClient

	gotConfig     *container.Config
	gotHostConfig *container.HostConfig
	gotNetConfig  *network.NetworkingConfig
	gotPlatform   *v1.Platform
	gotName       string

	resp container.CreateResponse
	err  error
}

func (f *fakeCreateAPI) ContainerCreate(
	_ context.Context,
	config *container.Config,
	hostConfig *container.HostConfig,
	networkingConfig *network.NetworkingConfig,
	platform *v1.Platform,
	containerName string,
) (container.CreateResponse, error) {
	f.gotConfig = config
	f.gotHostConfig = hostConfig
	f.gotNetConfig = networkingConfig
	f.gotPlatform = platform
	f.gotName = containerName

	return f.resp, f.err
}

func newLabelRuntime(api client.APIClient, suffix string, bakedLabels []string) *labelBasedRuntime {
	common := commonRuntime{
		cli: api,
	}

	return &labelBasedRuntime{
		commonRuntime: common,
		suffix:        suffix,
		bakedLabels:   bakedLabels,
	}
}

func newCreateRequest(name string) ContainerCreateRequest {
	cfg := &container.Config{
		Image: "img:v1",
	}
	hostCfg := &container.HostConfig{}
	netCfg := &network.NetworkingConfig{}
	platform := &v1.Platform{}

	return ContainerCreateRequest{
		Config:           &ContainerConfig{Config: cfg},
		HostConfig:       &HostConfig{HostConfig: hostCfg},
		NetworkingConfig: &NetworkingConfig{NetworkingConfig: netCfg},
		Platform:         platform,
		ContainerName:    name,
	}
}

// The pre-multi-environment default: an empty suffix leaves the Docker
// container name byte-for-byte what it always was.
func TestLabelBasedRuntime_EmptySuffixLeavesNameUntouched(t *testing.T) {
	api := &fakeCreateAPI{}
	runtime := newLabelRuntime(api, "", nil)

	_, err := runtime.ContainerCreate(context.Background(), newCreateRequest(testSmerdName))
	require.NoError(t, err)

	require.Equal(t, testSmerdName, api.gotName)
	require.Empty(t, api.gotConfig.Labels[labels.SuffixLabel])
}

// Name-conflict resolution: with a non-empty suffix the actual Docker
// container name carries it, so two environments deploying the same logical
// name don't collide on the shared daemon.
func TestLabelBasedRuntime_NonEmptySuffixSuffixesName(t *testing.T) {
	api := &fakeCreateAPI{}
	runtime := newLabelRuntime(api, testSuffix, nil)

	_, err := runtime.ContainerCreate(context.Background(), newCreateRequest(testSmerdName))
	require.NoError(t, err)

	require.Equal(t, testSmerdName+"_"+testSuffix, api.gotName)
	require.Equal(t, testSuffix, api.gotConfig.Labels[labels.SuffixLabel])
}

// Label stamping must stay identical to docker.Docker.ContainerCreate's:
// CREATED_WITH_VELEZ, the suffix label, and the node's baked custom labels in
// both "name=value" and bare "name" form.
func TestLabelBasedRuntime_StampsVelezAndBakedLabels(t *testing.T) {
	api := &fakeCreateAPI{}
	runtime := newLabelRuntime(api, testSuffix, []string{"team=core", "bare"})

	req := newCreateRequest(testSmerdName)

	req.Config.Labels = map[string]string{"caller": "kept"}

	_, err := runtime.ContainerCreate(context.Background(), req)
	require.NoError(t, err)

	require.Equal(t, "true", api.gotConfig.Labels[labels.CreatedWithVelezLabel])
	require.Equal(t, testSuffix, api.gotConfig.Labels[labels.SuffixLabel])
	require.Equal(t, "core", api.gotConfig.Labels["team"])
	require.Empty(t, api.gotConfig.Labels["bare"])
	require.Equal(t, "kept", api.gotConfig.Labels["caller"])
}

// The wrappers must pass their embedded Docker SDK values through untouched.
func TestLabelBasedRuntime_PassesThroughDockerSdkValues(t *testing.T) {
	api := &fakeCreateAPI{}

	api.resp = container.CreateResponse{ID: "cont-id"}

	runtime := newLabelRuntime(api, "", nil)

	req := newCreateRequest(testSmerdName)

	created, err := runtime.ContainerCreate(context.Background(), req)
	require.NoError(t, err)

	require.Equal(t, "cont-id", created.ID)
	require.Same(t, req.Config.Config, api.gotConfig)
	require.Same(t, req.HostConfig.HostConfig, api.gotHostConfig)
	require.Same(t, req.NetworkingConfig.NetworkingConfig, api.gotNetConfig)
	require.Same(t, req.Platform, api.gotPlatform)
}

// A missing config is a caller bug, not something to paper over with an
// implicit empty one - the Docker API would reject it anyway, less clearly.
func TestLabelBasedRuntime_MissingConfigErrors(t *testing.T) {
	api := &fakeCreateAPI{}
	runtime := newLabelRuntime(api, "", nil)

	req := ContainerCreateRequest{
		ContainerName: testSmerdName,
	}

	_, err := runtime.ContainerCreate(context.Background(), req)
	require.Error(t, err)
	require.Empty(t, api.gotName, "docker must not be called")
}

// Nil host/networking config is legal (docker treats them as defaults) and
// must not be dereferenced by the wrapper unwrapping.
func TestLabelBasedRuntime_NilOptionalConfigsArePassedAsNil(t *testing.T) {
	api := &fakeCreateAPI{}
	runtime := newLabelRuntime(api, "", nil)

	cfg := &container.Config{}

	req := ContainerCreateRequest{
		Config:        &ContainerConfig{Config: cfg},
		ContainerName: testSmerdName,
	}

	_, err := runtime.ContainerCreate(context.Background(), req)
	require.NoError(t, err)

	require.Nil(t, api.gotHostConfig)
	require.Nil(t, api.gotNetConfig)
}
