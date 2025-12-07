package network_steps

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
)

type issueClientKeyStep struct {
	networkService cluster_clients.VervClosedNetworkClient

	namespaceId *string

	keyResponse *string
}

func IssueClientKey(
	vpnClient cluster_clients.VervClosedNetworkClient,
	namespaceId *string,
	keyResponse *string,
) *issueClientKeyStep {
	return &issueClientKeyStep{
		networkService: vpnClient,
		namespaceId:    namespaceId,

		keyResponse: keyResponse,
	}
}

func (h *issueClientKeyStep) Do(ctx context.Context) error {
	clientKey, err := h.networkService.IssueClientKey(ctx, *h.namespaceId)
	if err != nil {
		return rerrors.Wrap(err)
	}

	*h.keyResponse = clientKey
	return nil
}
