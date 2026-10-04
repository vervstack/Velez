package jobs

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/network_owner"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	stepDiscoverSidecars = "discover_sidecars"
)

var upgradeLeftoverSuffixes = []string{
	oldContainerSuffix,
	replacedContainerSuffix,
	newContainerSuffix,
	configFetcherContainerSuffix,
}

type sidecarNamesAccessor interface {
	GetIsSidecarsSkipped() bool
	SetSidecarContainerNames(v []string)
}

// upgradeSidecarNames names the sidecars of rootId, leaving out the scratch
// containers an upgrade itself leaves behind.
func upgradeSidecarNames(list []container.Summary, rootId string) []string {
	names := make([]string, 0)

	for _, sidecar := range network_owner.SidecarsOf(list, rootId) {
		name := summaryContainerName(sidecar)
		if name == "" || hasAnySuffix(name, upgradeLeftoverSuffixes) {
			continue
		}

		names = append(names, name)
	}

	return names
}

func hasAnySuffix(name string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}

	return false
}

// discoverSidecarsJob records the containers sharing the old container's
// network namespace, before the upgrade replaces it.
type discoverSidecarsJob struct {
	runtimes container_runtime.RuntimeResolver

	upgradeReq upgradeRequestAccessor
	old        oldContainerIDAccessor
	ctx        sidecarNamesAccessor
}

func (j *discoverSidecarsJob) Do(ctx context.Context) error {
	if j.ctx.GetIsSidecarsSkipped() {
		return nil
	}

	runtime, err := j.runtimes.Runtime(ctx, j.upgradeReq.GetUpgradeRequest().GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	list, err := runtime.ListAllContainers(ctx, listAllContainersLimitNone)
	if err != nil {
		return rerrors.Wrap(err, "error listing containers")
	}

	j.ctx.SetSidecarContainerNames(upgradeSidecarNames(list, j.old.GetOldContainerId()))

	return nil
}

// upgradeSidecarGroup adapts an upgrade payload to the group
// recreate_sidecars works on: the upgraded container is the network root.
type upgradeSidecarGroup struct {
	payload *velez_api.UpgradeSmerdTaskPayload
}

func (g upgradeSidecarGroup) GetEnvironment() string {
	return g.payload.GetUpgradeRequest().GetEnvironment()
}

func (g upgradeSidecarGroup) GetServiceName() string {
	return g.payload.GetRequest().GetLabels()[labels.VervServiceLabel]
}

func (g upgradeSidecarGroup) GetContainerName() string {
	return g.payload.GetUpgradeRequest().GetName()
}

func (g upgradeSidecarGroup) GetSidecarContainerNames() []string {
	if g.payload.GetIsSidecarsSkipped() {
		return nil
	}

	return g.payload.GetSidecarContainerNames()
}
