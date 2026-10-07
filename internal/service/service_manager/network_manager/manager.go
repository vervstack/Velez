package network_manager

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/user_errors"
)

type ClusterModeProbe interface {
	IsClusterMode() bool
}

type NetworkManager struct {
	runtimes    container_runtime.RuntimeResolver
	vcn         cluster_clients.VervClosedNetworkClient
	clusterMode ClusterModeProbe
}

func New(
	runtimes container_runtime.RuntimeResolver,
	vcn cluster_clients.VervClosedNetworkClient,
	clusterMode ClusterModeProbe,
) *NetworkManager {
	return &NetworkManager{
		runtimes:    runtimes,
		vcn:         vcn,
		clusterMode: clusterMode,
	}
}

func (m *NetworkManager) GetStatus(
	ctx context.Context, _ *velez_api.GetNetworkStatus_Request,
) (*velez_api.GetNetworkStatus_Response, error) {
	_, err := m.vcn.ListNamespaces(ctx)
	isVcnConnected := err == nil

	providers := []*velez_api.NetworkProviderInfo{dockerProviderInfo()}
	if isVcnConnected {
		providers = append(providers, vcnProviderInfo())
	}

	return &velez_api.GetNetworkStatus_Response{
		IsClusterMode:  m.clusterMode.IsClusterMode(),
		IsVcnConnected: isVcnConnected,
		Providers:      providers,
	}, nil
}

func (m *NetworkManager) ListNetworks(
	ctx context.Context, req *velez_api.ListNetworks_Request,
) (*velez_api.ListNetworks_Response, error) {
	runtime, err := m.runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment")
	}

	infos, err := runtime.ListNetworks(ctx, req.GetIsForeignIncluded())
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing networks")
	}

	networks := make([]*velez_api.Network, 0, len(infos))
	for _, info := range infos {
		networks = append(networks, toProtoNetwork(info))
	}

	return &velez_api.ListNetworks_Response{Networks: networks}, nil
}

func (m *NetworkManager) GetNetwork(
	ctx context.Context, req *velez_api.GetNetwork_Request,
) (*velez_api.GetNetwork_Response, error) {
	runtime, err := m.runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment")
	}

	info, err := runtime.InspectNetwork(ctx, req.GetId())
	if err != nil {
		return nil, rerrors.Wrap(err, "error inspecting network")
	}

	return &velez_api.GetNetwork_Response{Network: toProtoNetwork(info)}, nil
}

func (m *NetworkManager) CreateNetwork(
	ctx context.Context, req *velez_api.CreateNetwork_Request,
) (*velez_api.CreateNetwork_Response, error) {
	name := strings.TrimSpace(req.GetName())
	if name == "" {
		return nil, rerrors.Wrap(user_errors.ErrNetworkNameEmpty)
	}

	runtime, err := m.runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment")
	}

	createReq := container_runtime.CreateNetworkRequest{
		Name:         name,
		IsInternal:   req.GetIsInternal(),
		IsIccEnabled: req.GetIsIccEnabled(),
	}

	info, err := runtime.CreateManagedNetwork(ctx, createReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating network")
	}

	return &velez_api.CreateNetwork_Response{Network: toProtoNetwork(info)}, nil
}

func (m *NetworkManager) DeleteNetwork(
	ctx context.Context, req *velez_api.DeleteNetwork_Request,
) (*velez_api.DeleteNetwork_Response, error) {
	runtime, err := m.runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment")
	}

	info, err := runtime.InspectNetwork(ctx, req.GetId())
	if err != nil {
		return nil, rerrors.Wrap(err, "error inspecting network")
	}

	if !info.IsManaged {
		return nil, rerrors.Wrap(user_errors.ErrNetworkNotManaged)
	}

	if len(info.Members) > 0 {
		return nil, rerrors.Wrap(user_errors.ErrNetworkNotEmpty)
	}

	err = runtime.RemoveNetwork(ctx, info.Id)
	if err != nil {
		return nil, rerrors.Wrap(err, "error removing network")
	}

	return &velez_api.DeleteNetwork_Response{}, nil
}

func (m *NetworkManager) ConnectContainer(
	ctx context.Context, req *velez_api.ConnectContainer_Request,
) (*velez_api.ConnectContainer_Response, error) {
	containerName := strings.TrimSpace(req.GetContainerName())
	if containerName == "" {
		return nil, rerrors.Wrap(user_errors.ErrNoSuchContainer)
	}

	runtime, err := m.runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment")
	}

	attachReq := container_runtime.AttachContainerRequest{
		ContainerID: containerName,
		NetworkId:   req.GetNetworkId(),
		Aliases:     req.GetAliases(),
	}

	err = runtime.AttachContainer(ctx, attachReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error connecting container")
	}

	return &velez_api.ConnectContainer_Response{}, nil
}

func (m *NetworkManager) DisconnectContainer(
	ctx context.Context, req *velez_api.DisconnectContainer_Request,
) (*velez_api.DisconnectContainer_Response, error) {
	containerName := strings.TrimSpace(req.GetContainerName())
	if containerName == "" {
		return nil, rerrors.Wrap(user_errors.ErrNoSuchContainer)
	}

	runtime, err := m.runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment")
	}

	err = runtime.DetachContainer(ctx, containerName, req.GetNetworkId())
	if err != nil {
		return nil, rerrors.Wrap(err, "error disconnecting container")
	}

	return &velez_api.DisconnectContainer_Response{}, nil
}
