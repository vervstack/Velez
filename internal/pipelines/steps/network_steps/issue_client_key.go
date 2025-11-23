package network_steps

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/service"
)

type issueClientKeyStep struct {
	networkService service.VervPrivateNetworkService

	namespaceId string

	keyResponse *string
}

func IssueClientKey(
	networkService service.VervPrivateNetworkService,
	namespace string,
	keyResponse *string,
) *issueClientKeyStep {
	return &issueClientKeyStep{
		networkService: networkService,
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
