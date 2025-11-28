package network_steps

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients"
)

type issueClientKeyStep struct {
	networkService clients.VervPrivateNetworkClient

	namespaceId string

	keyResponse *string
}

func IssueClientKey(
	clusterClients clients.ClusterClients,
	namespace string,
	keyResponse *string,
) *issueClientKeyStep {
	return &issueClientKeyStep{
		networkService: clusterClients.Vpn(),
		namespaceId:    namespace,

		keyResponse: keyResponse,
	}
}

func (h *issueClientKeyStep) Do(ctx context.Context) error {
	clientKey, err := h.networkService.IssueClientKey(ctx, h.namespaceId)
	if err != nil {
		return rerrors.Wrap(err)
	}

	*h.keyResponse = clientKey
	return nil
}
