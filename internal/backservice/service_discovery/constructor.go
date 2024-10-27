package service_discovery

import (
	"strconv"

	errors "github.com/Red-Sock/trace-errors"

	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/config"
)

func getPortToExposeTo(envVar config.EnvironmentConfig, nodeClients clients.NodeClients) (*string, error) {
	if !envVar.MakoshExposePort {
		return nil, nil
	}
	p := uint32(envVar.MakoshPort)

	if p == 0 {
		var err error
		p, err = nodeClients.PortManager().GetPort()
		if err != nil {
			return nil, errors.Wrap(err, "error obtaining port from pool")
		}
	}

	portStr := strconv.Itoa(int(p))

	return &portStr, nil
}

func getTargetURL(envVar config.EnvironmentConfig, internalClients clients.NodeClients, portToExposeTo *string) (string, error) {
	targetURL := Name

	isInContainer, err := env.IsInContainer(internalClients.Docker())
	if err != nil {
		return "", errors.Wrap(err, "can't determine if Velez is running inside a container")
	}

	if isInContainer {
		return targetURL, nil
	}

	if !envVar.MakoshExposePort {
		return "", errors.Wrap(ErrRequireMakoshPortExportToRunAsDaemon, "Makosh expose port is set to false")
	}
	if portToExposeTo == nil {
		return "", errors.Wrap(ErrRequireMakoshPortExportToRunAsDaemon, "No port to expose to")
	}

	targetURL = "0.0.0.0:" + *portToExposeTo

	return targetURL, nil
}
