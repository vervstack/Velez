package container_manager

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/storage/environments"
)

const (
	inspectTestSuffix = "stage"
	inspectTestName   = "mysvc"
	inspectTestImage  = "img:latest"
)

// fakeImageInspectAPI is a minimal client.APIClient fake for the one
// client.APIClient method InspectSmerd still calls directly - ImageInspect.
// Container inspection itself now goes through a resolved
// container_runtime.ContainerRuntime (fakeListContainerRuntime in
// smerd_list_test.go), not the raw dockerAPI - see InspectSmerd's doc
// comment in inspect.go.
type fakeImageInspectAPI struct {
	client.APIClient

	imgResp image.InspectResponse
	imgErr  error
}

func (f *fakeImageInspectAPI) ImageInspect(
	_ context.Context, _ string, _ ...client.ImageInspectOption,
) (image.InspectResponse, error) {
	return f.imgResp, f.imgErr
}

// newInspectContainerResponse builds a minimal-but-valid container.InspectResponse:
// enough non-nil fields for InspectSmerd to walk without a nil dereference.
// name is passed already in the virtual/logical form - the shape a real
// labelBasedRuntime.Inspect would have returned (see label_based_test.go's
// own TestLabelBasedRuntime_Inspect_FoundAndOwned_NameRewrittenToVirtual) -
// since suffix-stripping is no longer InspectSmerd's job.
func newInspectContainerResponse(virtualName string) container.InspectResponse {
	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:      "container-id",
			Name:    "/" + virtualName,
			Image:   "sha256:abc",
			Created: "2024-01-01T00:00:00Z",
			State:   &container.State{Status: "running"},
			HostConfig: &container.HostConfig{
				PortBindings: nil,
				Mounts:       nil,
			},
		},
		Config: &container.Config{
			Env:    nil,
			Labels: map[string]string{},
		},
		NetworkSettings: &container.NetworkSettings{},
	}
}

func newTestContainerManager(api client.APIClient, resolver *fakeRuntimeResolver) *ContainerManager {
	return &ContainerManager{
		dockerAPI: api,
		runtimes:  resolver,
	}
}

// InspectSmerd must pass contInfo.Name straight through (minus the leading
// "/") rather than re-deriving/re-stripping any suffix itself - that's the
// resolved ContainerRuntime's job now (label_based.go's Inspect).
func TestInspectSmerd_ReturnsRuntimeInspectResultAsSmerd(t *testing.T) {
	resolver := &fakeRuntimeResolver{
		envs:         environments.NewStatic([]string{inspectTestSuffix}, ""),
		inspectResp:  newInspectContainerResponse(inspectTestName),
		inspectFound: true,
	}
	api := &fakeImageInspectAPI{imgResp: image.InspectResponse{RepoTags: []string{inspectTestImage}}}

	smerd, err := newTestContainerManager(api, resolver).
		InspectSmerd(context.Background(), inspectTestSuffix, "container-id")
	require.NoError(t, err)
	require.Equal(t, inspectTestName, smerd.GetName())
	require.Equal(t, "container-id", smerd.GetUuid())
	require.Equal(t, inspectTestSuffix, resolver.gotEnvironment,
		"InspectSmerd must resolve the runtime for the given environment")
}

// A container not found (or not owned) by the resolved runtime is a real,
// surfaced error - not a nil Smerd.
func TestInspectSmerd_NotFound_ReturnsError(t *testing.T) {
	resolver := &fakeRuntimeResolver{
		envs:         environments.NewStatic(nil, ""),
		inspectFound: false,
	}
	api := &fakeImageInspectAPI{}

	smerd, err := newTestContainerManager(api, resolver).InspectSmerd(context.Background(), "", "ghost")
	require.Error(t, err)
	require.Nil(t, smerd)
}

// A failure to resolve the environment (unknown environment, storage error,
// etc.) propagates as an error.
func TestInspectSmerd_UnknownEnvironmentRejected(t *testing.T) {
	resolver := &fakeRuntimeResolver{envs: environments.NewStatic(nil, "")}
	api := &fakeImageInspectAPI{}

	smerd, err := newTestContainerManager(api, resolver).InspectSmerd(context.Background(), "ghost-env", "container-id")
	require.Error(t, err)
	require.Nil(t, smerd)
}

// An unexpected error from the resolved runtime's Inspect call propagates.
func TestInspectSmerd_RuntimeInspectErrorPropagates(t *testing.T) {
	resolver := &fakeRuntimeResolver{
		envs:       environments.NewStatic(nil, ""),
		inspectErr: rerrors.New("docker down"),
	}
	api := &fakeImageInspectAPI{}

	smerd, err := newTestContainerManager(api, resolver).InspectSmerd(context.Background(), "", "container-id")
	require.Error(t, err)
	require.Nil(t, smerd)
}
