package configuration

import (
	"strconv"

	errors "github.com/Red-Sock/trace-errors"

	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/config"
)

var (
	ErrRequireMatreshkaPortExportToRunAsDaemon = errors.New("matreshka port must be exported in order to run velez as daemon")
)

func getPort(cfg config.Config, cls clients.NodeClients) (*int, error) {
	envVars := cfg.Environment

	if !envVars.ExposeMatreshkaPort {
		return nil, nil
	}

	p := envVars.MatreshkaPort
	if p != 0 {
		return &p, nil
	}

	portFromManager, err := cls.PortManager().GetPort()
	if err != nil {
		return nil, errors.Wrap(err, "error obtaining port from pool")
	}

	p = int(portFromManager)

	return &p, nil

}

func getTargetURL(cfg config.Config, nodeClients clients.NodeClients, portToExposeTo *int) (string, error) {
	targetURL := Name

	isInContainer, err := env.IsInContainer(nodeClients.Docker())
	if err != nil {
		return "", errors.Wrap(err, "can't determine if Velez is running inside a container")
	}

	if isInContainer {
		return targetURL, nil
	}

	if !cfg.Environment.ExposeMatreshkaPort {
		return "", errors.Wrap(ErrRequireMatreshkaPortExportToRunAsDaemon, "Matreshka expose port is set to false")
	}

	if portToExposeTo == nil {
		return "", errors.Wrap(ErrRequireMatreshkaPortExportToRunAsDaemon, "No port to expose to")
	}

	return "http://0.0.0.0:" + strconv.Itoa(*portToExposeTo), nil
}
