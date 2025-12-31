package verv_closed_network

import (
	"context"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/domain"
)

type DisabledVcnImpl struct {
}

func (d DisabledVcnImpl) GetClientAuthKey(ctx context.Context, req domain.GetVcnAuthKeyReq) (domain.VcnAuthKey, error) {
	return domain.VcnAuthKey{}, cluster_clients.ErrServiceIsDisabled
}

func (d DisabledVcnImpl) RegisterNode(ctx context.Context, req domain.RegisterVcnNodeReq) error {
	return cluster_clients.ErrServiceIsDisabled
}

func (d DisabledVcnImpl) GetNamespace(ctx context.Context, name string) (domain.VcnNamespace, error) {
	return domain.VcnNamespace{}, cluster_clients.ErrServiceIsDisabled
}

func (d DisabledVcnImpl) CreateNamespace(_ context.Context, name string) (domain.VcnNamespace, error) {
	return domain.VcnNamespace{}, cluster_clients.ErrServiceIsDisabled
}

func (d DisabledVcnImpl) ListNamespaces(_ context.Context) ([]domain.VcnNamespace, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (d DisabledVcnImpl) DeleteNamespace(_ context.Context, id string) error {
	return cluster_clients.ErrServiceIsDisabled
}

func (d DisabledVcnImpl) IssueClientKey(_ context.Context, namespace domain.IssueClientKey) (string, error) {
	return "", cluster_clients.ErrServiceIsDisabled
}
