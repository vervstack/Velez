package jobs

import (
	"context"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

func inspectHaproxySidecar(
	ctx context.Context, runtimes container_runtime.RuntimeResolver, id, name string,
) (container.InspectResponse, error) {
	if id == "" {
		return container.InspectResponse{}, user_errors.New("Set up a VCN sidecar first")
	}

	runtime, err := runtimes.Runtime(ctx, "")
	if err != nil {
		return container.InspectResponse{}, rerrors.Wrap(err)
	}

	sidecar, isFound, err := runtime.Inspect(ctx, id)
	if err != nil {
		return container.InspectResponse{}, rerrors.Wrap(err)
	}

	if !isFound || sidecar.Config == nil || sidecar.State == nil || !sidecar.State.Running {
		return container.InspectResponse{}, user_errors.New("The VCN sidecar is unavailable")
	}

	owner := sidecar.Config.Labels[labels.HaproxySidecarLabel]
	if owner == "" || (name != "" && owner != name) {
		return container.InspectResponse{}, user_errors.New("Select the VCN sidecar created for this HAProxy instance")
	}

	return sidecar, nil
}

func connectHaproxyBackendNetworks(ctx context.Context, cli client.APIClient, id, networkIds string) error {
	if networkIds == "" {
		return nil
	}

	info, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting HAProxy network owner")
	}

	for networkId := range strings.SplitSeq(networkIds, ",") {
		if hasHaproxyNetwork(info, networkId) {
			continue
		}

		settings := &network.EndpointSettings{}

		err = cli.NetworkConnect(ctx, networkId, info.ID, settings)
		if err != nil {
			return rerrors.Wrap(err, "error connecting HAProxy backend network")
		}
	}

	return nil
}

func hasHaproxyNetwork(info container.InspectResponse, networkId string) bool {
	if info.NetworkSettings != nil {
		for _, endpoint := range info.NetworkSettings.Networks {
			if endpoint != nil && endpoint.NetworkID == networkId {
				return true
			}
		}
	}

	return false
}

func sidecarPublishesHaproxyPort(sidecar container.InspectResponse, port uint32) bool {
	if sidecar.HostConfig == nil {
		return false
	}

	key := nat.Port(strconv.FormatUint(uint64(port), 10) + "/tcp")

	return len(sidecar.HostConfig.PortBindings[key]) > 0
}

func connectHaproxyDefaultNetwork(
	ctx context.Context, runtime container_runtime.ContainerRuntime, cli client.APIClient,
	sidecar container.InspectResponse, alias string,
) error {
	err := runtime.CreateNetwork(ctx, env.VervNetwork)
	if err != nil {
		return rerrors.Wrap(err)
	}

	networks, err := runtime.ListNetworks(ctx, false)
	if err != nil {
		return rerrors.Wrap(err)
	}

	for _, info := range networks {
		if info.Name != env.VervNetwork || !info.IsManaged {
			continue
		}

		if hasHaproxyNetwork(sidecar, info.Id) {
			return nil
		}

		settings := &network.EndpointSettings{Aliases: []string{alias}}

		err = cli.NetworkConnect(ctx, info.Id, sidecar.ID, settings)
		if err != nil {
			return rerrors.Wrap(err)
		}

		return nil
	}

	return user_errors.New("The HAProxy environment network is unavailable")
}
