package dockerutils

import (
	"context"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/list_request"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	bridgeIccOption = "com.docker.network.bridge.enable_icc"
)

// CreateNetwork ensures a Velez-managed bridge network exists. No-op if a network with that exact name exists.
func CreateNetwork(ctx context.Context, d client.APIClient, networkName string) error {
	opts := CreateNetworkOptions{
		Labels:       map[string]string{labels.NetworkManagedLabel: labels.NetworkManagedLabelValue},
		IsIccEnabled: true,
	}

	_, err := CreateNetworkWithOptions(ctx, d, networkName, opts)
	if err != nil {
		return rerrors.Wrap(err, "error creating network")
	}

	return nil
}

type CreateNetworkOptions struct {
	Labels       map[string]string
	IsInternal   bool
	IsIccEnabled bool
}

// CreateNetworkWithOptions creates a bridge network and returns its id. If a network with that exact name
// already exists it is left untouched and its id is returned.
func CreateNetworkWithOptions(
	ctx context.Context,
	d client.APIClient,
	networkName string,
	opts CreateNetworkOptions,
) (id string, err error) {
	f := list_request.New()
	f.Name(networkName)

	listOpts := network.ListOptions{
		Filters: f.Args(),
	}

	existing, err := d.NetworkList(ctx, listOpts)
	if err != nil {
		return "", rerrors.Wrap(err, "error inspecting network for service")
	}

	for _, item := range existing {
		if item.Name == networkName {
			return item.ID, nil
		}
	}

	createOps := network.CreateOptions{
		Driver:   "bridge",
		Internal: opts.IsInternal,
		Labels:   opts.Labels,
	}

	if !opts.IsIccEnabled {
		createOps.Options = map[string]string{bridgeIccOption: "false"}
	}

	resp, err := d.NetworkCreate(ctx, networkName, createOps)
	if err != nil {
		return "", rerrors.Wrap(err, "error creating network for service")
	}

	return resp.ID, nil
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

func DisconnectFromNetworks(
	ctx context.Context,
	d client.APIClient,
	contId string,
) (disconnectedNetworks map[string]*network.EndpointSettings, err error) {
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
