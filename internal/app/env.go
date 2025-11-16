package app

import (
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/backservice/env"
)

func (c *Custom) setupVervNodeEnvironment() (err error) {
	// Verv network for communication inside node
	err = env.StartNetwork(c.NodeClients.Docker().Client())
	if err != nil {
		return rerrors.Wrap(err, "error creating network")
	}

	// Verv volumes for persistence inside node
	err = env.StartVolumes(c.NodeClients.Docker().Client())
	if err != nil {
		return rerrors.Wrap(err, "error creating volumes")
	}

	return nil
}
