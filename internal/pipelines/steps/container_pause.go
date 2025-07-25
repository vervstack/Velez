package steps

import (
	"context"
	"strconv"

	errdefs2 "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
)

type detachContainerFromVervStep struct {
	docker      clients.Docker
	portManager clients.PortManager

	req         domain.LaunchSmerd
	containerId *string

	disconnectedNets map[string]*network.EndpointSettings

	portsOnHold      []uint32
	stateBeforePause container.ContainerState
}

func PauseContainer(
	nodeClients clients.NodeClients,
	containerId *string,
) *detachContainerFromVervStep {
	return &detachContainerFromVervStep{
		docker:      nodeClients.Docker(),
		portManager: nodeClients.PortManager(),

		containerId: containerId,
	}
}

func (s *detachContainerFromVervStep) Do(ctx context.Context) error {
	if s.containerId == nil {
		return rerrors.New("container id is required")
	}

	cont, err := s.docker.InspectContainer(ctx, *s.containerId)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	s.stateBeforePause = cont.State.Status

	err = s.stopContainer(ctx, cont)
	if err != nil {
		return rerrors.Wrap(err, "error stopping container")
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

func (s *detachContainerFromVervStep) Rollback(ctx context.Context) error {
	if s.containerId == nil {
		return nil
	}

	if s.stateBeforePause != container.StateRunning {
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

func (s *detachContainerFromVervStep) stopContainer(ctx context.Context, cont container.InspectResponse) error {
	switch cont.State.Status {
	case container.StateRunning:
		// Running. Can softly pause
		err := s.docker.ContainerPause(ctx, *s.containerId)
		if err != nil {
			if !errdefs2.IsConflict(err) {
				return rerrors.Wrap(err, "error pausing container")
			}
		}
	case container.StateCreated:
	//	Do nothing. Container created but not running
	case container.StatePaused:
	//	Do nothing. Already paused
	case container.StateRestarting:
		stopOps := container.StopOptions{}
		err := s.docker.ContainerStop(ctx, *s.containerId, stopOps)
		if err != nil {
			return rerrors.Wrap(err, "error stopping container")
		}
	case container.StateRemoving:
	// Do nothing. Container soon will be deleted
	case container.StateExited:
	// Do nothing. Container already stopped
	case container.StateDead:
		// Do nothing. Container already stopped
	}

	return nil
}
