package smerd_steps

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
)

type dropContainer struct {
	docker node_clients.Docker
	contId *string
}

func DropContainerStep(nodeClients node_clients.NodeClients, contId *string) *dropContainer {
	return &dropContainer{
		docker: nodeClients.Docker(),
		contId: contId,
	}
}

func (d *dropContainer) Do(ctx context.Context) error {
	if d.contId == nil {
		return nil
	}

	err := d.docker.Remove(ctx, *d.contId)
	if err != nil {
		return rerrors.Wrap(err, "error dropping container")
	}

	return nil
}
