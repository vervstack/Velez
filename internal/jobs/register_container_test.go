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

func Test_RegisteredContainerSpec_CarriesElevatedAccess(t *testing.T) {
	t.Parallel()

	info := container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			HostConfig: &container.HostConfig{Privileged: true, CapAdd: []string{"NET_ADMIN"}},
		},
	}

	spec := registeredContainerSpec("app", "", testRegisterServiceName, info)

	access := spec.GetSettings().GetElevatedAccess()
	require.True(t, access.GetIsPrivileged())
	require.Equal(t, []string{"NET_ADMIN"}, access.GetCapAdd())
}

func Test_RegisteredContainerSpec_PlainContainerHasNoElevatedAccess(t *testing.T) {
	t.Parallel()

	info := container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			HostConfig: &container.HostConfig{},
		},
	}

	spec := registeredContainerSpec("app", "", testRegisterServiceName, info)

	require.Nil(t, spec.GetSettings().GetElevatedAccess())
}

func Test_IsServiceMountsLinked(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		isStatefull    bool
		rootId         string
		previousRootId string
		want           bool
	}{
		{"cluster plain", true, "", "", true},
		{"cluster with role change", true, "root", "previous", true},
		{"single-node plain", false, "", "", true},
		{"single-node root set", false, "root", "", false},
		{"single-node previous root set", false, "", "previous", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, isServiceMountsLinked(tc.isStatefull, tc.rootId, tc.previousRootId))
		})
	}
}
