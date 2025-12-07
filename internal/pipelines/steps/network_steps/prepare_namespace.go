package network_steps

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

type prepareNamespace struct {
	vcnClient     cluster_clients.VervPrivateNetworkClient
	namespaceName *string

	namespaceIdResp *string
}

func PrepareNamespace(
	vcnClient cluster_clients.VervPrivateNetworkClient,
	namespaceName *string,
	namespaceIdResp *string,
) steps.Step {
	return &prepareNamespace{
		vcnClient:       vcnClient,
		namespaceName:   namespaceName,
		namespaceIdResp: namespaceIdResp,
	}
}

func (s *prepareNamespace) Do(ctx context.Context) error {
	namespace, err := s.vcnClient.GetNamespace(ctx, *s.namespaceName)
	if err != nil {
		return rerrors.Wrap(err, "error getting vcs namespace when preparing")
	}

	if namespace.Id == "" {
		namespace, err = s.vcnClient.CreateNamespace(ctx, *s.namespaceName)
		if err != nil {
			return rerrors.Wrap(err, "error creating vcs namespace when preparing")
		}
	}

	*s.namespaceIdResp = namespace.Id

	return nil
}
