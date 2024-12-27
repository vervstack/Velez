package dockerutils

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils/list_request"
)

func CreateNetwork(ctx context.Context, d client.CommonAPIClient, networkName string) error {
	f := list_request.New()
	f.Name(networkName)
	net, err := d.NetworkList(ctx, types.NetworkListOptions{
		Filters: f.Args(),
	})
	if err != nil {
		return rerrors.Wrap(err, "error inspecting network for service")
	}

	for _, item := range net {
		if item.Name == networkName {
			return nil
		}
	}

	_, err = d.NetworkCreate(ctx,
		networkName,
		network.CreateOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error creating network for service")
	}

	return nil
}

type ConnectToNetworkRequest struct {
	NetworkName, ContId string
	Aliases             []string
}

func ConnectToNetwork(ctx context.Context, d client.CommonAPIClient, req ConnectToNetworkRequest) error {
	cont, err := d.ContainerInspect(ctx, req.ContId)
	if err != nil {
		return rerrors.Wrap(err, "error getting Velez container info")
	}

	isNetworkConnected := cont.NetworkSettings != nil && cont.NetworkSettings.Networks[req.NetworkName] != nil

	if !isNetworkConnected {
		connection := &network.EndpointSettings{
			Aliases: req.Aliases,
		}
		err = d.NetworkConnect(ctx, req.NetworkName, cont.Name, connection)
		if err != nil {
			return rerrors.Wrap(err, "error connecting this instance to network")
		}
	}

	return nil
}
