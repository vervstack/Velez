package env

import (
	"context"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
)

const (
	VervNetwork       = "verv"
	VelezNetworkAlias = "velez"
)

func StartNetwork(dockerAPI client.CommonAPIClient) error {
	ctx := context.Background()
	err := dockerutils.CreateNetwork(ctx, dockerAPI, VervNetwork)
	if err != nil {
		return rerrors.Wrap(err, "error creating network")
	}

	contId := GetContainerId()
	if contId == nil {
		return nil
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
