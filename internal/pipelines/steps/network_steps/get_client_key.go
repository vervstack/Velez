package network_steps

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/headscale"
	"go.vervstack.ru/Velez/internal/domain"
)

type getClientKeyStep struct {
	networkService cluster_clients.VervClosedNetworkClient

	namespaceId *string

	keyResponse *string
}

// GetClientKey - issues new client key if not exists in Network
// or returns existing
func GetClientKey(
	vpnClient cluster_clients.VervClosedNetworkClient,
	namespaceId *string,
	keyResponse *string,
) *getClientKeyStep {
	return &getClientKeyStep{
		networkService: vpnClient,
		namespaceId:    namespaceId,

		keyResponse: keyResponse,
	}
}

func (h *getClientKeyStep) Do(ctx context.Context) error {
	getAuthKeyReq := domain.GetVcnAuthKeyReq{
		NamespaceId:  *h.namespaceId,
		ReusableOnly: true,
	}
	authKey, err := h.networkService.GetClientAuthKey(ctx, getAuthKeyReq)
	if err != nil {
		if !rerrors.Is(err, headscale.ErrNotFound) {
			return rerrors.Wrap(err)
		}
	}

	if authKey.Key != "" {
		h.keyResponse = &authKey.Key
		return nil
	}

	issueClientKeyReq := domain.IssueClientKey{
		NamespaceId: *h.namespaceId,
		Reusable:    true,
	}

	clientKey, err := h.networkService.IssueClientKey(ctx, issueClientKeyReq)
	if err != nil {
		return rerrors.Wrap(err)
	}

	*h.keyResponse = clientKey
	return nil
}
