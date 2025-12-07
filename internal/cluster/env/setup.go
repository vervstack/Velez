package env

import (
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
)

func SetupEnvironment(clients node_clients.NodeClients) (err error) {
	// Verv network for communication inside node
	//err = StartNetwork(clients.Docker().Client())
	//if err != nil {
	//	return rerrors.Wrap(err, "error creating network")
	//}

	// Verv volumes for persistence inside node
	err = StartVolumes(clients.Docker().Client())
	if err != nil {
		return rerrors.Wrap(err, "error creating volumes")
	}

	return nil
}
