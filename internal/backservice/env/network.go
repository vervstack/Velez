package env

import (
	"context"
	"strings"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/clients/docker"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
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

	contId := GetContainerId()
	if contId == nil {
		return nil
	}

	_, err = dockerAPI.ContainerInspect(ctx, *contId)
	if err != nil {
		if strings.Contains(err.Error(), docker.NoSuchContainerError) {
			return nil
		}

		return rerrors.Wrap(err, "error inspecting container")
	}

	connReq := dockerutils.ConnectToNetworkRequest{
		NetworkName: VervNetwork,
		ContId:      *contId,
		Aliases:     []string{VelezNetworkAlias},
	}
	err = dockerutils.ConnectToNetwork(ctx, dockerAPI, connReq)
	if err != nil {
		return rerrors.Wrap(err, "error connecting velez to network")
	}

	return nil
}
