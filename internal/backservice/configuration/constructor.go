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

func getPort(cfg config.Config, cls clients.InternalClients) (*int, error) {
	envVars := cfg.Environment

	if !envVars.ExposeMatreshkaPort {
		return nil, nil
	}
	var p int

	p = envVars.MatreshkaPort
	if p == 0 {
		portFromManager, err := cls.PortManager().GetPort()
		if err != nil {
			return nil, errors.Wrap(err, "error obtaining port from pool")
		}

		p = int(portFromManager)
	}

	return &p, nil

}

func getTargetURL(cfg config.Config, internalClients clients.InternalClients, portToExposeTo *int) (string, error) {
	targetURL := Name
	envVar := cfg.Environment

	isInContainer, err := env.IsInContainer(internalClients.Docker())
	if err != nil {
		return "", errors.Wrap(err, "can't determine if Velez is running inside a container")
	}

	if !isInContainer {
		if !envVar.MakoshExposePort {
			return "", errors.Wrap(ErrRequireMatreshkaPortExportToRunAsDaemon, "Makosh expose port is set to false")
		}
		if portToExposeTo == nil {
			return "", errors.Wrap(ErrRequireMatreshkaPortExportToRunAsDaemon, "No port to expose to")
		}

		targetURL = "http://0.0.0.0:" + strconv.Itoa(*portToExposeTo)
	}

	return targetURL, nil
}
