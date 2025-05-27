package steps

import (
	"context"
	"strconv"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
)

type pauseContainerStep struct {
	docker      clients.Docker
	portManager clients.PortManager

	req         domain.LaunchSmerd
	containerId *string

	disconnectedNets map[string]*network.EndpointSettings

	portsOnHold []uint32
}

func PauseContainer(
	nodeClients clients.NodeClients,
	containerId *string,
) *pauseContainerStep {
	return &pauseContainerStep{
		docker:      nodeClients.Docker(),
		portManager: nodeClients.PortManager(),

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

	for _, hostPorts := range cont.NetworkSettings.Ports {
		for _, hostPort := range hostPorts {
			port, _ := strconv.ParseUint(hostPort.HostPort, 10, 32)
			p := uint32(port)
			s.portManager.HoldPort(p)
			s.portsOnHold = append(s.portsOnHold, p)
		}
	}

	return nil
}

func (s *pauseContainerStep) Rollback(ctx context.Context) error {
	if s.containerId == nil {
		return nil
	}

	for _, p := range s.portsOnHold {
		s.portManager.UnHoldPort(p)
	}

	err := s.docker.ContainerUnpause(ctx, *s.containerId)
	if err != nil {
		return rerrors.Wrapf(err, "error unpausing container '%s'", *s.containerId)
	}

	var globErr error

	for netName, net := range s.disconnectedNets {
		connReq := dockerutils.ConnectToNetworkRequest{
			NetworkName: netName,
			ContId:      *s.containerId,
			Aliases:     net.Aliases,
		}
		err = dockerutils.ConnectToNetwork(ctx, s.docker, connReq)
		if err != nil {
			globErr = rerrors.Join(globErr, rerrors.Wrap(err, "error connecting to network on rollback"))
		}
	}

	if globErr != nil {
		return rerrors.Wrap(globErr)
	}

	return nil
}
