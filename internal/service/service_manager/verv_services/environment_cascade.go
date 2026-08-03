package verv_services

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/list_request"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

// environmentContainerAPI is the narrow docker slice the cascade needs to find
// and remove an environment's containers. node_clients.Docker satisfies it.
//
// Removal deliberately goes through the very same Docker.Remove that
// container_manager.DropSmerds / internal/jobs' dropContainerJob use, and
// keeps their contract: a per-container failure is collected and reported,
// never silently swallowed, but it also doesn't abort the rest of the sweep.
type environmentContainerAPI interface {
	ListContainers(
		ctx context.Context, req *velez_api.ListSmerds_Request, suffix string,
	) ([]container.Summary, error)
	Remove(ctx context.Context, uuid string) error
}

// environmentNetworkVolumeAPI is the narrow docker slice the cascade needs for
// the non-container resources. *client.Client (client.APIClient) satisfies it.
type environmentNetworkVolumeAPI interface {
	NetworkList(ctx context.Context, options network.ListOptions) ([]network.Summary, error)
	NetworkRemove(ctx context.Context, networkID string) error
	VolumeList(ctx context.Context, options volume.ListOptions) (volume.ListResponse, error)
	VolumeRemove(ctx context.Context, volumeID string, force bool) error
}

func (v *VervService) cascadeRemoveEnvironmentResources(ctx context.Context, suffix string) error {
	return cascadeRemoveEnvironmentResources(ctx, v.docker, v.docker.Client(), suffix)
}

// cascadeRemoveEnvironmentResources stops/removes every container, network and
// volume tagged with labels.SuffixLabel == suffix.
//
// An empty suffix is refused on purpose: labels.SuffixLabel is written on
// every Velez-created container (as "" when no suffix is configured), so a
// blank filter would match - and destroy - the whole node.
func cascadeRemoveEnvironmentResources(
	ctx context.Context,
	containers environmentContainerAPI,
	netVol environmentNetworkVolumeAPI,
	suffix string,
) error {
	if suffix == "" {
		return rerrors.New("refusing to cascade-delete resources for an environment with an empty suffix")
	}

	err := removeEnvironmentContainers(ctx, containers, suffix)
	if err != nil {
		return err
	}

	err = removeEnvironmentNetworks(ctx, netVol, suffix)
	if err != nil {
		return err
	}

	return removeEnvironmentVolumes(ctx, netVol, suffix)
}

func removeEnvironmentContainers(ctx context.Context, api environmentContainerAPI, suffix string) error {
	req := &velez_api.ListSmerds_Request{
		Label: map[string]string{labels.SuffixLabel: suffix},
	}

	conts, err := api.ListContainers(ctx, req, suffix)
	if err != nil {
		return rerrors.Wrap(err, "error listing environment's containers")
	}

	var firstErr error

	for _, cont := range conts {
		// Docker.Remove is force-remove and already treats "no such
		// container" as success, so a running container needs no separate
		// stop call - exactly how DropSmerd removes a single smerd.
		removeErr := api.Remove(ctx, cont.ID)
		if removeErr != nil && firstErr == nil {
			firstErr = rerrors.Wrapf(removeErr, "error removing container '%s'", cont.ID)
		}
	}

	return firstErr
}

func removeEnvironmentNetworks(ctx context.Context, api environmentNetworkVolumeAPI, suffix string) error {
	filter := list_request.New()
	filter.Label(labels.SuffixLabel + "=" + suffix)

	listOpts := network.ListOptions{
		Filters: filter.Args(),
	}

	nets, err := api.NetworkList(ctx, listOpts)
	if err != nil {
		return rerrors.Wrap(err, "error listing environment's networks")
	}

	var firstErr error

	for _, net := range nets {
		removeErr := api.NetworkRemove(ctx, net.ID)
		if removeErr != nil && firstErr == nil {
			firstErr = rerrors.Wrapf(removeErr, "error removing network '%s'", net.ID)
		}
	}

	return firstErr
}

func removeEnvironmentVolumes(ctx context.Context, api environmentNetworkVolumeAPI, suffix string) error {
	filter := list_request.New()
	filter.Label(labels.SuffixLabel + "=" + suffix)

	listOpts := volume.ListOptions{
		Filters: filter.Args(),
	}

	vols, err := api.VolumeList(ctx, listOpts)
	if err != nil {
		return rerrors.Wrap(err, "error listing environment's volumes")
	}

	var firstErr error

	for _, vol := range vols.Volumes {
		if vol == nil {
			continue
		}

		removeErr := api.VolumeRemove(ctx, vol.Name, true)
		if removeErr != nil && firstErr == nil {
			firstErr = rerrors.Wrapf(removeErr, "error removing volume '%s'", vol.Name)
		}
	}

	return firstErr
}
