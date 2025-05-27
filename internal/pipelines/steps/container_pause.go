package steps

import (
	"context"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
)

type pauseContainerStep struct {
	docker clients.Docker

	req         domain.LaunchSmerd
	containerId *string

	disconnectedNets map[string]*network.EndpointSettings
}

func PauseContainer(
	nodeClients clients.NodeClients,
	containerId *string,
) *pauseContainerStep {
	return &pauseContainerStep{
		docker:      nodeClients.Docker(),
		containerId: containerId,
	}
}

func (s *pauseContainerStep) Do(ctx context.Context) error {
	if s.containerId == nil {
		return rerrors.New("container id is required")
	}

	err := s.docker.ContainerPause(ctx, *s.containerId)
	if err != nil {
		if !errdefs.IsConflict(err) {
			return rerrors.Wrap(err, "error pausing container")
		}
	}

	cont, err := s.docker.InspectContainer(ctx, *s.containerId)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	s.disconnectedNets, err = dockerutils.DisconnectFromNetworks(ctx, s.docker, cont.ID)
	if err != nil {
		return rerrors.Wrap(err, "error disconnecting from network")
	}

	return nil
}

func (s *pauseContainerStep) Rollback(ctx context.Context) error {
	if s.containerId == nil {
		return nil
	}

	err := s.docker.ContainerUnpause(ctx, *s.containerId)
	if err != nil {
		return rerrors.Wrapf(err, "error unpausing container '%s'", s.containerId)
	}

	for netName, net := range s.disconnectedNets {
		connReq := dockerutils.ConnectToNetworkRequest{
			NetworkName: netName,
			ContId:      *s.containerId,
			Aliases:     net.Aliases,
		}
		err = dockerutils.ConnectToNetwork(ctx, s.docker, connReq)
		if err != nil {
			return rerrors.Wrap(err, "error connecting to network on rollback")
		}
	}

	return nil
}
