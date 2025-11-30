package configuration

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients/matreshka"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/config"
)

func SetupMatreshka(
	ctx context.Context,
	cfg config.Config,
	nodeClients node_clients.NodeClients,
) (matreshka.Client, error) {

	if cfg.Environment.MasterNodeAddress != "" {
		// TODO implement across cluster connection
		return connectExternal(nodeClients)
	}

	client, err := deployOnThisNode(ctx, cfg, nodeClients)
	if err != nil {
		return nil, rerrors.Wrap(err, "error during node deployment")
	}

	return client, nil
}
