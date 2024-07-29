package service_discovery_task

import (
	"os"
	"strconv"

	rtb "github.com/Red-Sock/toolbox"
	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/makosh/pkg/makosh_resolver"

	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/config"
)

func getPortToExposeTo(envVar config.EnvironmentConfig, internalClients clients.InternalClients) (*string, error) {
	var portToExposeTo *string
	var err error
	if envVar.MakoshExposePort {
		p := uint32(envVar.MakoshPort)

		if p == 0 {
			p, err = internalClients.PortManager().GetPort()
			if err != nil {
				return nil, errors.Wrap(err, "error obtaining port from pool")
			}
		}

		portStr := strconv.Itoa(int(p))
		portToExposeTo = &portStr
	}

	return portToExposeTo, nil
}

func getTargetURL(envVar config.EnvironmentConfig, internalClients clients.InternalClients, portToExposeTo *string) (string, error) {
	targetURL := Name

	isInContainer, err := env.IsInContainer(internalClients.Docker())
	if err != nil {
		return "", errors.Wrap(err, "can't determine if Velez is running inside a container")
	}

	if !isInContainer {
		if !envVar.MakoshExposePort {
			return "", errors.Wrap(ErrRequireMakoshPortExportToRunAsDaemon, "Makosh expose port is set to false")
		}
		if portToExposeTo == nil {
			return "", errors.Wrap(ErrRequireMakoshPortExportToRunAsDaemon, "No port to expose to")
		}

		targetURL = "http://0.0.0.0:" + *portToExposeTo
	}

	err = os.Setenv(makosh_resolver.MakoshURL, targetURL)
	if err != nil {
		return "", errors.Wrap(err, "error setting environment variable for makosh target")
	}

	return targetURL, nil
}

func generateAuthToken() (string, error) {
	authToken, err := rtb.Random(256)
	if err != nil {
		return "", errors.Wrap(err, "error generating random token")
	}
	out := string(authToken)
	err = os.Setenv(makosh_resolver.MakoshSecret, out)
	if err != nil {
		return "", errors.Wrap(err, "error setting environment variable for makosh secret")
	}

	return out, nil
}
