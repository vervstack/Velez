package dockerutils

import (
	"context"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/list_request"
)

func CreateNetwork(ctx context.Context, d client.APIClient, networkName string) error {
	f := list_request.New()
	f.Name(networkName)

	req := network.ListOptions{
		Filters: f.Args(),
	}

	net, err := d.NetworkList(ctx, req)
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

func ConnectToNetwork(ctx context.Context, d client.APIClient, req ConnectToNetworkRequest) error {
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

func DisconnectFromNetworks(ctx context.Context, d client.APIClient, contId string) (disconnectedNetworks map[string]*network.EndpointSettings, err error) {
	cont, err := d.ContainerInspect(ctx, contId)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting Velez container info")
	}

	disconnectedNetworks = make(map[string]*network.EndpointSettings)

	for netName, net := range cont.NetworkSettings.Networks {
		err = d.NetworkDisconnect(ctx, netName, cont.Name, false)
		if err != nil {
			return nil, rerrors.Wrap(err, "error connecting this instance to network")
		}

		disconnectedNetworks[netName] = net
	}

	return disconnectedNetworks, nil
}
