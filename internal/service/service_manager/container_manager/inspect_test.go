package container_manager

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	inspectTestSuffix = "stage"
	inspectTestName   = "mysvc"
	inspectTestImage  = "img:latest"
)

// fakeInspectAPI is a minimal client.APIClient fake for InspectSmerd - only
// ContainerInspect and ImageInspect are exercised.
type fakeInspectAPI struct {
	client.APIClient

	contResp container.InspectResponse
	contErr  error

	imgResp image.InspectResponse
	imgErr  error
}

func (f *fakeInspectAPI) ContainerInspect(_ context.Context, _ string) (container.InspectResponse, error) {
	return f.contResp, f.contErr
}

func (f *fakeInspectAPI) ImageInspect(
	_ context.Context, _ string, _ ...client.ImageInspectOption,
) (image.InspectResponse, error) {
	return f.imgResp, f.imgErr
}

// newInspectContainerResponse builds a minimal-but-valid container.InspectResponse:
// enough non-nil fields for InspectSmerd to walk without a nil dereference.
func newInspectContainerResponse(dockerName, suffix string) container.InspectResponse {
	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:      "container-id",
			Name:    dockerName,
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
			Labels: map[string]string{labels.SuffixLabel: suffix},
		},
		NetworkSettings: &container.NetworkSettings{},
	}
}

func newTestContainerManager(api *fakeInspectAPI) *ContainerManager {
	return &ContainerManager{
		dockerAPI: api,
	}
}

// InspectSmerd must surface the virtual/logical name, not the suffixed real
// Docker name - the CreateSmerd response Name leak this test guards against
// (see docs/container_runtimes/interface_design.md).
func TestInspectSmerd_StripsSuffixFromName(t *testing.T) {
	api := &fakeInspectAPI{
		contResp: newInspectContainerResponse("/"+inspectTestName+"_"+inspectTestSuffix, inspectTestSuffix),
		imgResp:  image.InspectResponse{RepoTags: []string{inspectTestImage}},
	}

	smerd, err := newTestContainerManager(api).InspectSmerd(context.Background(), "container-id")
	require.NoError(t, err)
	require.Equal(t, inspectTestName, smerd.GetName())
}

// An empty suffix (the pre-multi-environment default) leaves the name
// untouched, byte for byte.
func TestInspectSmerd_EmptySuffixLeavesNameUntouched(t *testing.T) {
	api := &fakeInspectAPI{
		contResp: newInspectContainerResponse("/"+inspectTestName, ""),
		imgResp:  image.InspectResponse{RepoTags: []string{inspectTestImage}},
	}

	smerd, err := newTestContainerManager(api).InspectSmerd(context.Background(), "container-id")
	require.NoError(t, err)
	require.Equal(t, inspectTestName, smerd.GetName())
}

// A container with no SuffixLabel at all (containers predating the
// multi-environment feature) must not have anything stripped - suffix
// defaults to "" and StripEnvironmentSuffix is a no-op for that.
func TestInspectSmerd_NoSuffixLabel_LeavesNameUntouched(t *testing.T) {
	resp := newInspectContainerResponse("/"+inspectTestName, "")

	resp.Config.Labels = map[string]string{}

	api := &fakeInspectAPI{
		contResp: resp,
		imgResp:  image.InspectResponse{RepoTags: []string{inspectTestImage}},
	}

	smerd, err := newTestContainerManager(api).InspectSmerd(context.Background(), "container-id")
	require.NoError(t, err)
	require.Equal(t, inspectTestName, smerd.GetName())
}
