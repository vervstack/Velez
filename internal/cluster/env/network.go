package env

import (
	"context"
	"strings"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
)

const (
	VervNetwork       = "verv"
	VelezNetworkAlias = "velez"
)

func StartNetwork(dockerAPI client.APIClient) error {
	ctx := context.Background()

	err := dockerutils.CreateNetwork(ctx, dockerAPI, VervNetwork)
	if err != nil {
		return rerrors.Wrap(err, "error creating network")
	}

	contID := GetContainerId()
	if contID == nil {
		return nil
	}

	_, err = dockerAPI.ContainerInspect(ctx, *contID)
	if err != nil {
		if strings.Contains(err.Error(), docker.NoSuchContainerError) {
			return nil
		}

		return rerrors.Wrap(err, "error inspecting container")
	}

	connReq := dockerutils.ConnectToNetworkRequest{
		NetworkName: VervNetwork,
		ContId:      *contID,
		Aliases:     []string{VelezNetworkAlias},
	}

	err = dockerutils.ConnectToNetwork(ctx, dockerAPI, connReq)
	if err != nil {
		return rerrors.Wrap(err, "error connecting velez to network")
	}

	return nil
}
