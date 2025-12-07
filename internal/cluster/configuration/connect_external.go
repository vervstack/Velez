package configuration

import (
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/app/matreshka_client"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients/matreshka"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
)

func newClient(nodeClients node_clients.NodeClients) (matreshka.Client, error) {
	matreshkaClient, err := matreshka.NewClient(
		grpc.WithUnaryInterceptor(
			matreshka_client.WithHeader(
				matreshka_client.Pass, nodeClients.LocalStateManager().Get().MatreshkaKey)))
	if err != nil {
		return nil, rerrors.Wrap(err, "error initializing matreshka client ")
	}

	return matreshkaClient, nil
}
