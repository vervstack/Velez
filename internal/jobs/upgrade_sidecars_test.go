package jobs

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

func newUpgradeSummary(id, name, networkMode string) container.Summary {
	summary := container.Summary{
		ID:    id,
		Names: []string{"/" + name},
	}

	summary.HostConfig.NetworkMode = networkMode

	return summary
}

func Test_UpgradeSidecarNames_SkipsUpgradeLeftovers(t *testing.T) {
	t.Parallel()

	list := []container.Summary{
		newUpgradeSummary("aaa111", "app", "bridge"),
		newUpgradeSummary("bbb222", "ts", "container:app"),
		newUpgradeSummary("ccc333", "ts_old", "container:app"),
		newUpgradeSummary("ddd444", "ts_new", "container:app"),
		newUpgradeSummary("eee555", "ts_configuration_fetcher", "container:app"),
		newUpgradeSummary("fff666", "other", "bridge"),
	}

	require.Equal(t, []string{"ts"}, upgradeSidecarNames(list, "aaa111"))
}

func Test_UpgradeSidecarNames_NoSidecars(t *testing.T) {
	t.Parallel()

	list := []container.Summary{newUpgradeSummary("aaa111", "app", "bridge")}

	require.Empty(t, upgradeSidecarNames(list, "aaa111"))
}

func Test_UpgradeSidecarGroup_ReadsPayload(t *testing.T) {
	t.Parallel()

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: "app", Environment: "prod"},
		Request: &velez_api.CreateSmerd_Request{
			Labels: map[string]string{labels.VervServiceLabel: "svc"},
		},
	}
	group := upgradeSidecarGroup{payload: payload}

	require.Empty(t, group.GetSidecarContainerNames())

	payload.SetSidecarContainerNames([]string{"ts"})

	require.Equal(t, "prod", group.GetEnvironment())
	require.Equal(t, "svc", group.GetServiceName())
	require.Equal(t, "app", group.GetContainerName())
	require.Equal(t, []string{"ts"}, group.GetSidecarContainerNames())
}

func Test_UpgradeSidecarGroup_UnlabeledRootHasEmptyServiceName(t *testing.T) {
	t.Parallel()

	payload := &velez_api.UpgradeSmerdTaskPayload{
		Request: &velez_api.CreateSmerd_Request{},
	}

	require.Empty(t, upgradeSidecarGroup{payload: payload}.GetServiceName())
}

func Test_UpgradeSidecarGroup_SkippedHidesSidecars(t *testing.T) {
	t.Parallel()

	payload := &velez_api.UpgradeSmerdTaskPayload{
		SidecarContainerNames: []string{"ts"},
		IsSidecarsSkipped:     true,
	}

	require.Empty(t, upgradeSidecarGroup{payload: payload}.GetSidecarContainerNames())
}

func Test_SidecarLabels_EmptyServiceOnlyMarksSidecar(t *testing.T) {
	t.Parallel()

	require.Equal(t, map[string]string{labels.Sidecar: labelTrueValue}, sidecarLabels(""))
}
