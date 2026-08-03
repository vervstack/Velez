package container_runtime

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
)

// commonRuntime backs operations that don't differ by backend. It only needs a
// raw client.APIClient and doesn't know - or care - whether that connection
// points at the node's shared daemon or at one dedicated to a single
// environment.
//
// CreateNetwork/ConnectToNetwork/DisconnectFromNetworks are NOT here despite
// being backend-agnostic in principle: docs/container_runtimes/roadmap.md's
// Stage 4 design tied them to labelBasedRuntime instead, since a network name
// needs the same suffix translation container names get.
type commonRuntime struct {
	cli client.APIClient
}

// PullImage mirrors docker.Docker.PullImage exactly (same dockerutils.PullImage
// call, same follow-up ImageInspect) - duplicated rather than delegated to
// avoid commonRuntime depending on the node_clients.Docker wrapper type.
func (c *commonRuntime) PullImage(ctx context.Context, imageName string) (image.InspectResponse, error) {
	_, err := dockerutils.PullImage(ctx, c.cli, imageName, false)
	if err != nil {
		return image.InspectResponse{}, rerrors.Wrap(err, "error pulling image")
	}

	img, err := c.cli.ImageInspect(ctx, imageName)
	if err != nil {
		return image.InspectResponse{}, rerrors.Wrap(err, "error inspecting image")
	}

	return img, nil
}

// ListOccupiedPorts mirrors docker.Docker.ListOccupiedPorts exactly: every
// public port bound by any container on the daemon, unfiltered by suffix - see
// this method's doc comment on the ContainerRuntime interface for why that's
// correct here specifically.
func (c *commonRuntime) ListOccupiedPorts(ctx context.Context) ([]uint32, error) {
	containerList, err := c.cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	usedPorts := make([]uint32, 0)

	for _, cont := range containerList {
		for _, p := range cont.Ports {
			if p.PublicPort != 0 {
				usedPorts = append(usedPorts, uint32(p.PublicPort))
			}
		}
	}

	return usedPorts, nil
}
