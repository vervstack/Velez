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
	"go.vervstack.ru/Velez/tests/test_helper"
)

const (
	testSmerdName          = "mysvc"
	testSuffix             = "stage"
	testExecEchoCmd        = "echo hello"
	testBakedLabelTeamCore = "team=core"
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

func newLabelRuntime(api client.APIClient, suffix string, bakedLabels []string) *dockerRuntime {
	common := commonRuntime{
		cli: api,
	}

	resolver := &labelSuffixResolver{
		suffix: suffix,
	}

	return &dockerRuntime{
		commonRuntime: common,
		resolver:      resolver,
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

// newLabelBasedRuntime must build a dockerRuntime backed by a
// labelSuffixResolver carrying the given suffix - the production constructor
// resolver.go's label-based branch calls.
func TestNewLabelBasedRuntime_BuildsLabelSuffixResolverBackedRuntime(t *testing.T) {
	api := &fakeCreateAPI{}

	runtime := newLabelBasedRuntime(api, testSuffix, []string{testBakedLabelTeamCore})

	resolver, ok := runtime.resolver.(*labelSuffixResolver)
	require.True(t, ok)
	require.Equal(t, testSuffix, resolver.suffix)
	require.Equal(t, []string{testBakedLabelTeamCore}, runtime.bakedLabels)
}

// newDirectRuntime must build a dockerRuntime backed by a directResolver -
// tier 2's constructor, not yet wired into production (resolver.go still
// returns ErrDedicatedRuntimeNotImplemented for a dedicated environment).
func TestNewDirectRuntime_BuildsDirectResolverBackedRuntime(t *testing.T) {
	api := &fakeCreateAPI{}

	runtime := newDirectRuntime(api, []string{testBakedLabelTeamCore})

	_, ok := runtime.resolver.(*directResolver)
	require.True(t, ok)
	require.Equal(t, []string{testBakedLabelTeamCore}, runtime.bakedLabels)
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
	runtime := newLabelRuntime(api, testSuffix, []string{testBakedLabelTeamCore, "bare"})

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

// Rename and IsContainerRunning are exercised against a real local Docker
// daemon - see createRealContainer below - per the project's "no
// hand-written mocks/fakes, prefer real API calls" testing policy
// (CLAUDE.md's "Testing" section). fakeRemoveAPI/newFakeRemoveAPI above are
// untouched: they back the separate ContainerCreate/ListContainers/Remove
// families, deferred to a later pass.

// createRealContainer builds a dockerRuntime for suffix/bakedLabels and
// creates a real container through that runtime's own already-tested
// ContainerCreate (reusing already-verified production code as the fixture
// path, rather than a second hand-rolled create call), using
// test_helper.HelloWorldAppImage. Returns the container's real Docker ID and
// registers t.Cleanup removal.
func createRealContainer(
	t *testing.T,
	api client.APIClient,
	suffix string,
	bakedLabels []string,
	name string,
) string {
	t.Helper()

	test_helper.EnsurePulled(t, api, test_helper.HelloWorldAppImage)

	runtime := newLabelRuntime(api, suffix, bakedLabels)

	req := ContainerCreateRequest{
		Config:        &ContainerConfig{Config: &container.Config{Image: test_helper.HelloWorldAppImage}},
		ContainerName: name,
	}

	created, err := runtime.ContainerCreate(context.Background(), req)
	require.NoError(t, err)

	t.Cleanup(func() {
		test_helper.RemoveContainer(t, api, created.ID)
	})

	return created.ID
}

// A bare logical name in a suffixed environment must resolve via the
// suffixed Docker name (the exact resolution Remove already does), and the
// requested new name must itself be run through containerName() before
// being handed to Docker - the actual fix for today's separate bug where
// upgrade_smerd.go's rename steps apply no suffix at all to the renamed
// container.
func TestLabelBasedRuntime_Rename_BareNameResolvesAndSuffixesNewName(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)

	name := test_helper.UniqueName(t, testSmerdName)
	newName := test_helper.UniqueName(t, "newname")

	createRealContainer(t, api, testSuffix, nil, name)

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.Rename(context.Background(), name, newName)
	require.NoError(t, err)

	inspected, err := api.ContainerInspect(context.Background(), runtime.resolver.ContainerName(newName))
	require.NoError(t, err, "expected the renamed, suffixed container to exist on the daemon")
	require.Equal(t, testSuffix, inspected.Config.Labels[labels.SuffixLabel])
}

// A raw Docker UUID never matches the suffixed form, so resolution must fall
// back to the identifier exactly as given - same as Remove.
func TestLabelBasedRuntime_Rename_UuidFallsBackToRawIdentifier(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)

	name := test_helper.UniqueName(t, testSmerdName)
	newName := test_helper.UniqueName(t, "newname")

	id := createRealContainer(t, api, testSuffix, nil, name)

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.Rename(context.Background(), id, newName)
	require.NoError(t, err)

	inspected, err := api.ContainerInspect(context.Background(), runtime.resolver.ContainerName(newName))
	require.NoError(t, err)
	require.Equal(t, id, inspected.ID)
}

// Suffix-mismatch (container exists but belongs to a different environment)
// is treated as not-found / idempotent success - no rename, no error - same
// cross-environment protection as Remove.
func TestLabelBasedRuntime_Rename_SuffixMismatch_TreatedAsNotFound(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)

	otherSuffix := test_helper.UniqueName(t, "othersfx")
	name := test_helper.UniqueName(t, testSmerdName)
	newName := test_helper.UniqueName(t, "newname")

	otherRuntime := newLabelRuntime(api, otherSuffix, nil)
	id := createRealContainer(t, api, otherSuffix, nil, name)

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.Rename(context.Background(), id, newName)
	require.NoError(t, err, "a container belonging to a different environment must not be renamed")

	inspected, err := api.ContainerInspect(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, "/"+otherRuntime.resolver.ContainerName(name), inspected.Name,
		"the container's real Docker name must be unchanged")
}

// Not-found under either identifier form is idempotent success, same as
// Remove's identical case. No fixture needed at all.
func TestLabelBasedRuntime_Rename_NotFoundUnderEitherForm_IsIdempotentSuccess(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	runtime := newLabelRuntime(api, testSuffix, nil)

	name := test_helper.UniqueName(t, testSmerdName)

	err := runtime.Rename(context.Background(), name, "newname")
	require.NoError(t, err)
}

// A running container owned by this environment reports (true, true, nil).
func TestLabelBasedRuntime_IsContainerRunning_RunningOwnedContainer(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	name := test_helper.UniqueName(t, testSmerdName)

	id := createRealContainer(t, api, testSuffix, nil, name)

	err := api.ContainerStart(context.Background(), id, container.StartOptions{})
	require.NoError(t, err)

	runtime := newLabelRuntime(api, testSuffix, nil)

	running, exists, err := runtime.IsContainerRunning(context.Background(), name)
	require.NoError(t, err)
	require.True(t, exists)
	require.True(t, running)
}

// A stopped container owned by this environment reports (false, true, nil).
// A freshly created-but-never-started container is already in Docker's
// "created" (non-running) state, so no extra Stop/Pause call is needed to
// exercise this branch.
func TestLabelBasedRuntime_IsContainerRunning_StoppedOwnedContainer(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	name := test_helper.UniqueName(t, testSmerdName)

	createRealContainer(t, api, testSuffix, nil, name)

	runtime := newLabelRuntime(api, testSuffix, nil)

	running, exists, err := runtime.IsContainerRunning(context.Background(), name)
	require.NoError(t, err)
	require.True(t, exists)
	require.False(t, running)
}

// A container that doesn't exist under either identifier form reports
// (false, false, nil). No fixture needed at all.
func TestLabelBasedRuntime_IsContainerRunning_NotFound(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	runtime := newLabelRuntime(api, testSuffix, nil)

	name := test_helper.UniqueName(t, testSmerdName)

	running, exists, err := runtime.IsContainerRunning(context.Background(), name)
	require.NoError(t, err)
	require.False(t, exists)
	require.False(t, running)
}

// A container belonging to a different environment's suffix is treated
// exactly like "doesn't exist" - (false, false, nil) - same cross-environment
// protection as Remove/Rename.
func TestLabelBasedRuntime_IsContainerRunning_SuffixMismatch_TreatedAsNotFound(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)

	otherSuffix := test_helper.UniqueName(t, "othersfx")
	name := test_helper.UniqueName(t, testSmerdName)

	id := createRealContainer(t, api, otherSuffix, nil, name)

	runtime := newLabelRuntime(api, testSuffix, nil)

	running, exists, err := runtime.IsContainerRunning(context.Background(), id)
	require.NoError(t, err)
	require.False(t, exists)
	require.False(t, running)
}

// Inspect is exercised against a real local Docker daemon, same as
// Rename/IsContainerRunning above - see createRealContainer.

// A container owned by this environment (suffix matches) is found, and its
// Name is rewritten to the virtual/logical name, never the suffixed real
// Docker name.
func TestLabelBasedRuntime_Inspect_FoundAndOwned_NameRewrittenToVirtual(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	name := test_helper.UniqueName(t, testSmerdName)

	id := createRealContainer(t, api, testSuffix, nil, name)

	runtime := newLabelRuntime(api, testSuffix, nil)

	info, found, err := runtime.Inspect(context.Background(), name)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, id, info.ID)
	require.Equal(t, "/"+name, info.Name, "Name must be the virtual/logical name, not the suffixed Docker name")
}

// A container belonging to a different environment's suffix is treated
// exactly like "doesn't exist" - (zero value, false, nil) - same
// cross-environment protection as Remove/Rename/IsContainerRunning.
func TestLabelBasedRuntime_Inspect_SuffixMismatch_TreatedAsNotFound(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)

	otherSuffix := test_helper.UniqueName(t, "othersfx")
	name := test_helper.UniqueName(t, testSmerdName)

	id := createRealContainer(t, api, otherSuffix, nil, name)

	runtime := newLabelRuntime(api, testSuffix, nil)

	_, found, err := runtime.Inspect(context.Background(), id)
	require.NoError(t, err)
	require.False(t, found)
}

// A container that doesn't exist under either identifier form reports
// (zero value, false, nil). No fixture needed at all.
func TestLabelBasedRuntime_Inspect_NotFoundUnderEitherForm(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	runtime := newLabelRuntime(api, testSuffix, nil)

	name := test_helper.UniqueName(t, testSmerdName)

	_, found, err := runtime.Inspect(context.Background(), name)
	require.NoError(t, err)
	require.False(t, found)
}

// Stop/Restart/Stats are exercised against a real local Docker daemon, same
// as Rename/IsContainerRunning/Inspect above - see createRealContainer.

// A running container owned by this environment is actually stopped.
func TestLabelBasedRuntime_Stop_OwnedContainer_StopsIt(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	name := test_helper.UniqueName(t, testSmerdName)

	id := createRealContainer(t, api, testSuffix, nil, name)

	err := api.ContainerStart(context.Background(), id, container.StartOptions{})
	require.NoError(t, err)

	runtime := newLabelRuntime(api, testSuffix, nil)

	err = runtime.Stop(context.Background(), name)
	require.NoError(t, err)

	inspected, err := api.ContainerInspect(context.Background(), id)
	require.NoError(t, err)
	require.False(t, inspected.State.Running, "container must be stopped after Stop")
}

// A container belonging to a different environment's suffix is treated as
// "nothing to stop" - idempotent success, not an error - and left untouched,
// same cross-environment protection as Remove/Rename.
func TestLabelBasedRuntime_Stop_SuffixMismatch_TreatedAsNotFound(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)

	otherSuffix := test_helper.UniqueName(t, "othersfx")
	name := test_helper.UniqueName(t, testSmerdName)

	id := createRealContainer(t, api, otherSuffix, nil, name)

	err := api.ContainerStart(context.Background(), id, container.StartOptions{})
	require.NoError(t, err)

	runtime := newLabelRuntime(api, testSuffix, nil)

	err = runtime.Stop(context.Background(), id)
	require.NoError(t, err, "a container belonging to a different environment must not be stopped")

	inspected, err := api.ContainerInspect(context.Background(), id)
	require.NoError(t, err)
	require.True(t, inspected.State.Running, "the other environment's container must be left running")
}

// Not-found under either identifier form is idempotent success. No fixture
// needed at all.
func TestLabelBasedRuntime_Stop_NotFoundUnderEitherForm_IsIdempotentSuccess(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	runtime := newLabelRuntime(api, testSuffix, nil)

	name := test_helper.UniqueName(t, testSmerdName)

	err := runtime.Stop(context.Background(), name)
	require.NoError(t, err)
}

// A running container owned by this environment is restarted (still running
// afterwards - ContainerRestart blocks until the container is back up).
func TestLabelBasedRuntime_Restart_OwnedContainer_RestartsIt(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	name := test_helper.UniqueName(t, testSmerdName)

	id := createRealContainer(t, api, testSuffix, nil, name)

	err := api.ContainerStart(context.Background(), id, container.StartOptions{})
	require.NoError(t, err)

	runtime := newLabelRuntime(api, testSuffix, nil)

	err = runtime.Restart(context.Background(), name)
	require.NoError(t, err)

	inspected, err := api.ContainerInspect(context.Background(), id)
	require.NoError(t, err)
	require.True(t, inspected.State.Running, "container must be running again after Restart")
}

// A container belonging to a different environment's suffix is treated as
// "nothing to restart" - idempotent success, not an error - same
// cross-environment protection as Stop/Remove/Rename.
func TestLabelBasedRuntime_Restart_SuffixMismatch_TreatedAsNotFound(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)

	otherSuffix := test_helper.UniqueName(t, "othersfx")
	name := test_helper.UniqueName(t, testSmerdName)

	id := createRealContainer(t, api, otherSuffix, nil, name)

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.Restart(context.Background(), id)
	require.NoError(t, err, "a container belonging to a different environment must not be restarted")
}

// Not-found under either identifier form is idempotent success. No fixture
// needed at all.
func TestLabelBasedRuntime_Restart_NotFoundUnderEitherForm_IsIdempotentSuccess(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	runtime := newLabelRuntime(api, testSuffix, nil)

	name := test_helper.UniqueName(t, testSmerdName)

	err := runtime.Restart(context.Background(), name)
	require.NoError(t, err)
}

// A running container owned by this environment returns real stats.
func TestLabelBasedRuntime_Stats_OwnedRunningContainer_ReturnsStats(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	name := test_helper.UniqueName(t, testSmerdName)

	id := createRealContainer(t, api, testSuffix, nil, name)

	err := api.ContainerStart(context.Background(), id, container.StartOptions{})
	require.NoError(t, err)

	runtime := newLabelRuntime(api, testSuffix, nil)

	stats, err := runtime.Stats(context.Background(), name)
	require.NoError(t, err)
	require.False(t, stats.StartedAt.IsZero(), "a running container must report a non-zero StartedAt")
}

// A container belonging to a different environment's suffix is a real error
// here - unlike Stop/Restart/Remove, there's no sensible zero-value success
// for "stats of a container that isn't mine."
func TestLabelBasedRuntime_Stats_SuffixMismatch_ReturnsError(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)

	otherSuffix := test_helper.UniqueName(t, "othersfx")
	name := test_helper.UniqueName(t, testSmerdName)

	id := createRealContainer(t, api, otherSuffix, nil, name)

	runtime := newLabelRuntime(api, testSuffix, nil)

	_, err := runtime.Stats(context.Background(), id)
	require.Error(t, err)
}

// Not-found under either identifier form is a real error. No fixture needed
// at all.
func TestLabelBasedRuntime_Stats_NotFoundUnderEitherForm_ReturnsError(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	runtime := newLabelRuntime(api, testSuffix, nil)

	name := test_helper.UniqueName(t, testSmerdName)

	_, err := runtime.Stats(context.Background(), name)
	require.Error(t, err)
}

// Exec is exercised against a real local Docker daemon, same as
// Stop/Restart/Stats above - see createRealContainer.
// test_helper.HelloWorldAppImage is busybox-based (has /bin/sh), so a plain
// "sh -c" command is enough to exercise Exec without a dedicated image.

// A running container owned by this environment actually runs cfg and
// returns its output.
func TestLabelBasedRuntime_Exec_OwnedRunningContainer_ReturnsOutput(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	name := test_helper.UniqueName(t, testSmerdName)

	id := createRealContainer(t, api, testSuffix, nil, name)

	err := api.ContainerStart(context.Background(), id, container.StartOptions{})
	require.NoError(t, err)

	runtime := newLabelRuntime(api, testSuffix, nil)

	execCfg := container.ExecOptions{
		Cmd:          []string{"sh", "-c", testExecEchoCmd},
		AttachStdout: true,
		AttachStderr: true,
	}

	out, err := runtime.Exec(context.Background(), name, execCfg)
	require.NoError(t, err)
	require.Contains(t, string(out), "hello")
}

// A container belonging to a different environment's suffix is a real error
// here, unlike Stop/Restart/Remove - same rationale as Stats: there's no
// sensible zero-value success for "exec output of a container that isn't
// mine."
func TestLabelBasedRuntime_Exec_SuffixMismatch_ReturnsError(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)

	otherSuffix := test_helper.UniqueName(t, "othersfx")
	name := test_helper.UniqueName(t, testSmerdName)

	id := createRealContainer(t, api, otherSuffix, nil, name)

	err := api.ContainerStart(context.Background(), id, container.StartOptions{})
	require.NoError(t, err)

	runtime := newLabelRuntime(api, testSuffix, nil)

	execCfg := container.ExecOptions{
		Cmd:          []string{"sh", "-c", testExecEchoCmd},
		AttachStdout: true,
	}

	_, err = runtime.Exec(context.Background(), id, execCfg)
	require.Error(t, err, "a container belonging to a different environment must not be exec'd into")
}

// Not-found under either identifier form is a real error. No fixture needed
// at all.
func TestLabelBasedRuntime_Exec_NotFoundUnderEitherForm_ReturnsError(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	runtime := newLabelRuntime(api, testSuffix, nil)

	name := test_helper.UniqueName(t, testSmerdName)

	execCfg := container.ExecOptions{
		Cmd:          []string{"sh", "-c", testExecEchoCmd},
		AttachStdout: true,
	}

	_, err := runtime.Exec(context.Background(), name, execCfg)
	require.Error(t, err)
}

// CreateNetwork/ConnectToNetwork/DisconnectFromNetworks are exercised against
// a real local Docker daemon, same as Stop/Restart/Stats/Exec above. Each
// test uses its own unique base network name (test_helper.UniqueName) rather
// than the literal "verv" production constant (env.VervNetwork), to avoid
// colliding with env.StartNetwork's real shared network - or other
// tests/processes - on the same daemon: the suffixing rule under test
// (networkName, byte-for-byte containerName's rule) is identical regardless
// of which logical name gets suffixed.

// An empty suffix leaves the network name byte-for-byte unchanged - the
// pre-multi-environment default, mirroring
// TestLabelBasedRuntime_EmptySuffixLeavesNameUntouched for containers.
func TestLabelBasedRuntime_CreateNetwork_EmptySuffix_CreatesUnsuffixedName(t *testing.T) {
	api := test_helper.NewRealDockerAPI(t)
	base := test_helper.UniqueName(t, "vervnet")

	runtime := newLabelRuntime(api, "", nil)

	err := runtime.CreateNetwork(context.Background(), base)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = api.NetworkRemove(context.Background(), base)
	})

	_, err = api.NetworkInspect(context.Background(), base, network.InspectOptions{})
	require.NoError(t, err, "expected a network named %q on the daemon", base)
}

// A non-empty suffix creates "<name>_<suffix>" - each environment gets its
// own dedicated network instead of every environment sharing one (the real
// behavior change docs/container_runtimes/interface_design.md and the plan
// call for).
func TestLabelBasedRuntime_CreateNetwork_NonEmptySuffix_CreatesSuffixedName(t *testing.T) {
	api := test_helper.NewRealDockerAPI(t)
	base := test_helper.UniqueName(t, "vervnet")

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.CreateNetwork(context.Background(), base)
	require.NoError(t, err)

	expectedName := base + nameSuffixSeparator + testSuffix

	t.Cleanup(func() {
		_ = api.NetworkRemove(context.Background(), expectedName)
	})

	_, err = api.NetworkInspect(context.Background(), expectedName, network.InspectOptions{})
	require.NoError(t, err, "expected a network named %q on the daemon", expectedName)

	// The bare/unsuffixed name must NOT have been created - suffixing must
	// actually happen, not just be a no-op passthrough.
	_, err = api.NetworkInspect(context.Background(), base, network.InspectOptions{})
	require.Error(t, err, "the bare/unsuffixed network name must not exist")
}

// CreateNetwork is idempotent: calling it twice for the same logical name
// must not error.
func TestLabelBasedRuntime_CreateNetwork_Idempotent(t *testing.T) {
	api := test_helper.NewRealDockerAPI(t)
	base := test_helper.UniqueName(t, "vervnet")

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.CreateNetwork(context.Background(), base)
	require.NoError(t, err)

	expectedName := base + nameSuffixSeparator + testSuffix

	t.Cleanup(func() {
		_ = api.NetworkRemove(context.Background(), expectedName)
	})

	err = runtime.CreateNetwork(context.Background(), base)
	require.NoError(t, err, "creating an already-existing network must not error")
}

// A container owned by this environment gets connected to the (suffixed)
// network, and Docker actually reports the connection.
func TestLabelBasedRuntime_ConnectToNetwork_OwnedContainer_Connects(t *testing.T) {
	api := test_helper.NewRealDockerAPI(t)
	name := test_helper.UniqueName(t, testSmerdName)
	base := test_helper.UniqueName(t, "vervnet")

	id := createRealContainer(t, api, testSuffix, nil, name)

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.CreateNetwork(context.Background(), base)
	require.NoError(t, err)

	expectedNetName := base + nameSuffixSeparator + testSuffix

	t.Cleanup(func() {
		_ = api.NetworkRemove(context.Background(), expectedNetName)
	})

	connReq := ConnectToNetworkRequest{
		ContainerID: name,
		NetworkName: base,
		Aliases:     []string{"myalias"},
	}

	err = runtime.ConnectToNetwork(context.Background(), connReq)
	require.NoError(t, err)

	inspected, err := api.ContainerInspect(context.Background(), id)
	require.NoError(t, err)
	require.Contains(t, inspected.NetworkSettings.Networks, expectedNetName,
		"expected the container to be connected to the suffixed network")
}

// A container belonging to a different environment's suffix must not be
// connected - same cross-environment protection as Stop/Restart/Remove -
// except here it's a real error, not idempotent success (same reasoning as
// Exec/Stats: there's no sensible "connected" outcome for a container that
// isn't mine).
func TestLabelBasedRuntime_ConnectToNetwork_SuffixMismatch_ReturnsError(t *testing.T) {
	api := test_helper.NewRealDockerAPI(t)

	otherSuffix := test_helper.UniqueName(t, "othersfx")
	name := test_helper.UniqueName(t, testSmerdName)
	base := test_helper.UniqueName(t, "vervnet")

	id := createRealContainer(t, api, otherSuffix, nil, name)

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.CreateNetwork(context.Background(), base)
	require.NoError(t, err)

	expectedNetName := base + nameSuffixSeparator + testSuffix

	t.Cleanup(func() {
		_ = api.NetworkRemove(context.Background(), expectedNetName)
	})

	connReq := ConnectToNetworkRequest{
		ContainerID: id,
		NetworkName: base,
	}

	err = runtime.ConnectToNetwork(context.Background(), connReq)
	require.Error(t, err, "a container belonging to a different environment must not be connected")
}

// Not-found under either identifier form is a real error. No fixture needed
// at all.
func TestLabelBasedRuntime_ConnectToNetwork_NotFoundUnderEitherForm_ReturnsError(t *testing.T) {
	api := test_helper.NewRealDockerAPI(t)
	runtime := newLabelRuntime(api, testSuffix, nil)

	name := test_helper.UniqueName(t, testSmerdName)
	base := test_helper.UniqueName(t, "vervnet")

	connReq := ConnectToNetworkRequest{
		ContainerID: name,
		NetworkName: base,
	}

	err := runtime.ConnectToNetwork(context.Background(), connReq)
	require.Error(t, err)
}

// A container owned by this environment gets disconnected from the
// (suffixed) network it was connected to.
func TestLabelBasedRuntime_DisconnectFromNetworks_OwnedContainer_Disconnects(t *testing.T) {
	api := test_helper.NewRealDockerAPI(t)
	name := test_helper.UniqueName(t, testSmerdName)
	base := test_helper.UniqueName(t, "vervnet")

	id := createRealContainer(t, api, testSuffix, nil, name)

	runtime := newLabelRuntime(api, testSuffix, nil)

	err := runtime.CreateNetwork(context.Background(), base)
	require.NoError(t, err)

	expectedNetName := base + nameSuffixSeparator + testSuffix

	t.Cleanup(func() {
		_ = api.NetworkRemove(context.Background(), expectedNetName)
	})

	connReq := ConnectToNetworkRequest{ContainerID: name, NetworkName: base}

	err = runtime.ConnectToNetwork(context.Background(), connReq)
	require.NoError(t, err)

	err = runtime.DisconnectFromNetworks(context.Background(), name, []string{base})
	require.NoError(t, err)

	inspected, err := api.ContainerInspect(context.Background(), id)
	require.NoError(t, err)
	require.NotContains(t, inspected.NetworkSettings.Networks, expectedNetName,
		"expected the container to be disconnected from the suffixed network")
}

// A container belonging to a different environment's suffix must not be
// touched - real error, same reasoning as ConnectToNetwork.
func TestLabelBasedRuntime_DisconnectFromNetworks_SuffixMismatch_ReturnsError(t *testing.T) {
	api := test_helper.NewRealDockerAPI(t)

	otherSuffix := test_helper.UniqueName(t, "othersfx")
	name := test_helper.UniqueName(t, testSmerdName)
	base := test_helper.UniqueName(t, "vervnet")

	id := createRealContainer(t, api, otherSuffix, nil, name)

	otherRuntime := newLabelRuntime(api, otherSuffix, nil)

	err := otherRuntime.CreateNetwork(context.Background(), base)
	require.NoError(t, err)

	otherNetName := base + nameSuffixSeparator + otherSuffix

	t.Cleanup(func() {
		_ = api.NetworkRemove(context.Background(), otherNetName)
	})

	connReq := ConnectToNetworkRequest{ContainerID: id, NetworkName: base}

	err = otherRuntime.ConnectToNetwork(context.Background(), connReq)
	require.NoError(t, err)

	runtime := newLabelRuntime(api, testSuffix, nil)

	err = runtime.DisconnectFromNetworks(context.Background(), id, []string{base})
	require.Error(t, err, "a container belonging to a different environment must not be disconnected")

	inspected, err := api.ContainerInspect(context.Background(), id)
	require.NoError(t, err)
	require.Contains(t, inspected.NetworkSettings.Networks, otherNetName,
		"the other environment's container must remain connected")
}

// Not-found under either identifier form is a real error. No fixture needed
// at all.
func TestLabelBasedRuntime_DisconnectFromNetworks_NotFoundUnderEitherForm_ReturnsError(t *testing.T) {
	api := test_helper.NewRealDockerAPI(t)
	runtime := newLabelRuntime(api, testSuffix, nil)

	name := test_helper.UniqueName(t, testSmerdName)
	base := test_helper.UniqueName(t, "vervnet")

	err := runtime.DisconnectFromNetworks(context.Background(), name, []string{base})
	require.Error(t, err)
}
