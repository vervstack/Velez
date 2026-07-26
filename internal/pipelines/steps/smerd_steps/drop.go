package smerd_steps

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
)

type dropContainer struct {
	docker node_clients.Docker
	contID *string
}

func DropContainerStep(nodeClients node_clients.NodeClients, contID *string) *dropContainer {
	return &dropContainer{
		docker: nodeClients.Docker(),
		contID: contID,
	}
}

func (d *dropContainer) Do(ctx context.Context) error {
	if d.contID == nil {
		return nil
	}

	err := d.docker.Remove(ctx, *d.contID)
	if err != nil {
		return rerrors.Wrap(err, "error dropping container")
	}

	return nil
}
