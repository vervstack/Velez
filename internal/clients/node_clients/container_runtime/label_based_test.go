package container_runtime

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
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

// fakeListAPI is a minimal client.APIClient fake for ListContainers.
type fakeListAPI struct {
	client.APIClient

	resp []container.Summary
	err  error
}

func (f *fakeListAPI) ContainerList(_ context.Context, _ container.ListOptions) ([]container.Summary, error) {
	return f.resp, f.err
}

// The suffix label filter must be set unconditionally - even an empty suffix
// is a real filter value now, not "skip filtering" (see
// docs/container_runtimes/interface_design.md's "Suffix filtering has no
// 'unscoped' escape hatch").
func TestLabelBasedRuntime_ListContainers_SetsSuffixLabelUnconditionally(t *testing.T) {
	for _, suffix := range []string{"", testSuffix} {
		t.Run("suffix="+suffix, func(t *testing.T) {
			api := &fakeListAPI{}
			runtime := newLabelRuntime(api, suffix, nil)

			req := &velez_api.ListSmerds_Request{}

			_, err := runtime.ListContainers(context.Background(), req)
			require.NoError(t, err)

			gotSuffix, ok := req.GetLabel()[labels.SuffixLabel]
			require.True(t, ok, "the suffix label filter key must always be present")
			require.Equal(t, suffix, gotSuffix)
		})
	}
}

// Every name in every returned container.Summary must have this runtime's
// suffix stripped back off - callers of ContainerRuntime never see a
// suffixed Docker name.
func TestLabelBasedRuntime_ListContainers_StripsSuffixFromNames(t *testing.T) {
	api := &fakeListAPI{
		resp: []container.Summary{
			{ID: "a", Names: []string{"/" + testSmerdName + "_" + testSuffix}},
		},
	}
	runtime := newLabelRuntime(api, testSuffix, nil)

	list, err := runtime.ListContainers(context.Background(), &velez_api.ListSmerds_Request{})
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "/"+testSmerdName, list[0].Names[0])
}

// An empty suffix leaves names untouched - byte for byte what
// pre-multi-environment ListContainers always returned.
func TestLabelBasedRuntime_ListContainers_EmptySuffixLeavesNamesUntouched(t *testing.T) {
	api := &fakeListAPI{
		resp: []container.Summary{{ID: "a", Names: []string{"/" + testSmerdName}}},
	}
	runtime := newLabelRuntime(api, "", nil)

	list, err := runtime.ListContainers(context.Background(), &velez_api.ListSmerds_Request{})
	require.NoError(t, err)
	require.Equal(t, "/"+testSmerdName, list[0].Names[0])
}

// fakeRemoveAPI is a minimal client.APIClient fake for Remove/
// resolveOwnedContainer: ContainerInspect keyed by the identifier the caller
// passed in, plus ContainerRemove call recording.
type fakeRemoveAPI struct {
	client.APIClient

	inspectResp map[string]container.InspectResponse
	inspectErr  map[string]error
	removeErr   map[string]error

	inspectCalls []string
	removeCalls  []string
}

func newFakeRemoveAPI() *fakeRemoveAPI {
	return &fakeRemoveAPI{
		inspectResp: map[string]container.InspectResponse{},
		inspectErr:  map[string]error{},
		removeErr:   map[string]error{},
	}
}

func (f *fakeRemoveAPI) ContainerInspect(_ context.Context, id string) (container.InspectResponse, error) {
	f.inspectCalls = append(f.inspectCalls, id)

	if err, ok := f.inspectErr[id]; ok {
		return container.InspectResponse{}, err
	}

	if resp, ok := f.inspectResp[id]; ok {
		return resp, nil
	}

	notFoundMsg := "Error response from daemon: " + docker.NoSuchContainerError + ": " + id

	return container.InspectResponse{}, rerrors.New(notFoundMsg)
}

func (f *fakeRemoveAPI) ContainerRemove(_ context.Context, id string, _ container.RemoveOptions) error {
	f.removeCalls = append(f.removeCalls, id)

	return f.removeErr[id]
}

func newInspectResponse(id, suffix string) container.InspectResponse {
	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{ID: id},
		Config:            &container.Config{Labels: map[string]string{labels.SuffixLabel: suffix}},
	}
}

// A bare logical name in a suffixed environment must resolve via the
// suffixed Docker name - the bare-name-silently-no-ops bug's fix.
func TestLabelBasedRuntime_Remove_BareNameResolvesViaSuffixedForm(t *testing.T) {
	api := newFakeRemoveAPI()

	suffixedName := testSmerdName + "_" + testSuffix

	api.inspectResp[suffixedName] = newInspectResponse("real-id", testSuffix)

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.Remove(context.Background(), testSmerdName)
	require.NoError(t, err)

	require.Equal(t, []string{suffixedName}, api.inspectCalls)
	require.Equal(t, []string{"real-id"}, api.removeCalls)
}

// A raw Docker UUID never matches the suffixed form, so resolution must fall
// back to the identifier exactly as given.
func TestLabelBasedRuntime_Remove_FallsBackToRawIdentifierForUuid(t *testing.T) {
	api := newFakeRemoveAPI()

	const uuid = "container-uuid"

	api.inspectResp[uuid] = newInspectResponse(uuid, testSuffix)

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.Remove(context.Background(), uuid)
	require.NoError(t, err)

	require.Equal(t, []string{uuid + "_" + testSuffix, uuid}, api.inspectCalls,
		"suffixed form must be tried first, then the raw identifier")
	require.Equal(t, []string{uuid}, api.removeCalls)
}

// The cross-environment collision fix: a container resolved under a
// different suffix (e.g. a UUID belonging to another environment) is treated
// as not found here - Remove must not touch it, even though Docker's
// ContainerInspect/ContainerRemove would happily match it by ID regardless
// of environment.
func TestLabelBasedRuntime_Remove_SuffixMismatch_TreatedAsNotFound(t *testing.T) {
	api := newFakeRemoveAPI()

	const uuid = "other-env-uuid"

	api.inspectResp[uuid] = newInspectResponse(uuid, "some-other-suffix")

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.Remove(context.Background(), uuid)
	require.NoError(t, err)
	require.Empty(t, api.removeCalls, "a container belonging to a different environment must not be removed")
}

// Empty suffix is a real value to match now, not a wildcard: a runtime
// resolved for the empty-suffix environment must still refuse to remove a
// container whose own suffix label is non-empty.
func TestLabelBasedRuntime_Remove_EmptySuffixRuntimeRejectsNonEmptySuffixContainer(t *testing.T) {
	api := newFakeRemoveAPI()

	const uuid = "stage-uuid"

	api.inspectResp[uuid] = newInspectResponse(uuid, testSuffix)

	runtime := newLabelRuntime(api, "", nil)

	err := runtime.Remove(context.Background(), uuid)
	require.NoError(t, err)
	require.Empty(t, api.removeCalls)
	require.Equal(t, []string{uuid}, api.inspectCalls,
		"empty suffix means containerName(identifier) == identifier, so only one candidate is tried")
}

// No container under either identifier form is the same idempotent-success
// case Remove has always had for an already-gone container.
func TestLabelBasedRuntime_Remove_NotFoundUnderEitherForm_IsIdempotentSuccess(t *testing.T) {
	api := newFakeRemoveAPI()
	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.Remove(context.Background(), testSmerdName)
	require.NoError(t, err)
	require.Empty(t, api.removeCalls)
}

// An inspect failure that ISN'T "no such container" (Docker unreachable,
// etc.) must propagate as a real error rather than being swallowed as
// not-found.
func TestLabelBasedRuntime_Remove_UnexpectedInspectErrorPropagates(t *testing.T) {
	api := newFakeRemoveAPI()

	suffixedName := testSmerdName + "_" + testSuffix

	api.inspectErr[suffixedName] = rerrors.New("connection refused")

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.Remove(context.Background(), testSmerdName)
	require.Error(t, err)
	require.Empty(t, api.removeCalls)
}

// A ContainerRemove failure that isn't "no such container" must propagate,
// matching docker.Docker.Remove's existing semantics.
func TestLabelBasedRuntime_Remove_ContainerRemoveErrorPropagates(t *testing.T) {
	api := newFakeRemoveAPI()

	suffixedName := testSmerdName + "_" + testSuffix

	api.inspectResp[suffixedName] = newInspectResponse("real-id", testSuffix)
	api.removeErr["real-id"] = rerrors.New("daemon exploded")

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.Remove(context.Background(), testSmerdName)
	require.Error(t, err)
}

// A ContainerRemove "no such container" (e.g. a race where the container
// disappeared between inspect and remove) is still idempotent success.
func TestLabelBasedRuntime_Remove_ContainerRemoveNoSuchContainer_IsIdempotentSuccess(t *testing.T) {
	api := newFakeRemoveAPI()

	suffixedName := testSmerdName + "_" + testSuffix

	api.inspectResp[suffixedName] = newInspectResponse("real-id", testSuffix)
	api.removeErr["real-id"] = rerrors.New("Error: " + docker.NoSuchContainerError + ": real-id")

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.Remove(context.Background(), testSmerdName)
	require.NoError(t, err)
}
