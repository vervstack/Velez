package jobs

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	testRegisterServiceName = "billing"
)

func Test_RegisteredServiceLabels_CarryServiceAndDisplayName(t *testing.T) {
	t.Parallel()

	got := registeredServiceLabels(testRegisterServiceName)

	require.Equal(t, map[string]string{
		labels.VervServiceLabel: testRegisterServiceName,
		labels.DisplayNameLabel: testRegisterServiceName,
	}, got)
}

func Test_RegisteredContainerSpec_MergesLabelsAndStampedWin(t *testing.T) {
	t.Parallel()

	info := container.InspectResponse{
		Config: &container.Config{
			Image: "app:1",
			Labels: map[string]string{
				"keep":                  "me",
				labels.DisplayNameLabel: "old display",
			},
		},
		ContainerJSONBase: &container.ContainerJSONBase{
			HostConfig: &container.HostConfig{},
		},
	}

	spec := registeredContainerSpec("app", "prod", testRegisterServiceName, info)

	require.Equal(t, "app", spec.GetName())
	require.Equal(t, "app:1", spec.GetImageName())
	require.Equal(t, "prod", spec.GetEnvironment())
	require.Equal(t, map[string]string{
		"keep":                  "me",
		labels.VervServiceLabel: testRegisterServiceName,
		labels.DisplayNameLabel: testRegisterServiceName,
	}, spec.GetLabels())
}

func Test_RegisteredContainerSpec_NilConfigStillStampsLabels(t *testing.T) {
	t.Parallel()

	info := container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			HostConfig: &container.HostConfig{},
		},
	}

	spec := registeredContainerSpec("app", "", testRegisterServiceName, info)

	require.Equal(t, testRegisterServiceName, spec.GetLabels()[labels.VervServiceLabel])
}
