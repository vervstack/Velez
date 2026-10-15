package transfer_out

import (
	"cmp"
	"context"
	"slices"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

const (
	hostNetworkName     = "host"
	noneNetworkName     = "none"
	containerModePrefix = "container:"
)

// findConsumers lists containers, other than the transferred ones, attached
// to a network the root is attached to.
func (s *Service) findConsumers(
	ctx context.Context,
	root container.InspectResponse,
	sidecars []container.InspectResponse,
) ([]domain.TransferConsumer, error) {
	moved := movedIds(root, sidecars)
	consumers := make([]domain.TransferConsumer, 0)

	for _, attached := range attachedNetworks(root) {
		inspected, err := s.docker.Client().NetworkInspect(ctx, attached.id, network.InspectOptions{})
		if err != nil {
			return nil, rerrors.Wrap(err, "error inspecting network")
		}

		consumers = append(consumers, networkConsumers(inspected, attached.name, moved)...)
	}

	slices.SortFunc(consumers, compareConsumers)

	return consumers, nil
}

func compareConsumers(a, b domain.TransferConsumer) int {
	return cmp.Or(
		cmp.Compare(a.ContainerName, b.ContainerName),
		cmp.Compare(a.NetworkName, b.NetworkName),
	)
}

type attachedNetwork struct {
	id   string
	name string
}

func attachedNetworks(root container.InspectResponse) []attachedNetwork {
	hasOwnNetwork := root.HostConfig != nil && !strings.HasPrefix(string(root.HostConfig.NetworkMode), containerModePrefix)
	if !hasOwnNetwork || root.NetworkSettings == nil {
		return nil
	}

	networks := make([]attachedNetwork, 0, len(root.NetworkSettings.Networks))

	for name, endpoint := range root.NetworkSettings.Networks {
		isShared := name != hostNetworkName && name != noneNetworkName
		if !isShared || endpoint == nil {
			continue
		}

		networks = append(networks, attachedNetwork{id: endpoint.NetworkID, name: name})
	}

	return networks
}

func networkConsumers(
	inspected network.Inspect,
	networkName string,
	moved map[string]struct{},
) []domain.TransferConsumer {
	consumers := make([]domain.TransferConsumer, 0, len(inspected.Containers))

	for id, endpoint := range inspected.Containers {
		_, isMoved := moved[id]
		if isMoved {
			continue
		}

		consumer := domain.TransferConsumer{
			ContainerName: endpoint.Name,
			NetworkName:   networkName,
		}

		consumers = append(consumers, consumer)
	}

	return consumers
}

func movedIds(root container.InspectResponse, sidecars []container.InspectResponse) map[string]struct{} {
	ids := map[string]struct{}{root.ID: {}}

	for _, sidecar := range sidecars {
		ids[sidecar.ID] = struct{}{}
	}

	return ids
}
