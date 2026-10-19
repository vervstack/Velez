package verv_closed_network

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

type DisabledVcnImpl struct{}

func (d DisabledVcnImpl) GetClientAuthKey(ctx context.Context, req domain.GetVcnAuthKeyReq) (domain.VcnAuthKey, error) {
	return domain.VcnAuthKey{}, rerrors.Wrap(user_errors.ErrServiceIsDisabled)
}

func (d DisabledVcnImpl) RegisterNode(ctx context.Context, req domain.RegisterVcnNodeReq) error {
	return rerrors.Wrap(user_errors.ErrServiceIsDisabled)
}

func (d DisabledVcnImpl) GetNamespace(ctx context.Context, name string) (domain.VcnNamespace, error) {
	return domain.VcnNamespace{}, rerrors.Wrap(user_errors.ErrServiceIsDisabled)
}

func (d DisabledVcnImpl) CreateNamespace(_ context.Context, name string) (domain.VcnNamespace, error) {
	return domain.VcnNamespace{}, rerrors.Wrap(user_errors.ErrServiceIsDisabled)
}

func (d DisabledVcnImpl) ListNamespaces(_ context.Context) ([]domain.VcnNamespace, error) {
	return nil, rerrors.Wrap(user_errors.ErrServiceIsDisabled)
}

func (d DisabledVcnImpl) DeleteNamespace(_ context.Context, id string) error {
	return rerrors.Wrap(user_errors.ErrServiceIsDisabled)
}

func (d DisabledVcnImpl) IssueClientKey(_ context.Context, namespace domain.IssueClientKey) (string, error) {
	return "", rerrors.Wrap(user_errors.ErrServiceIsDisabled)
}

func (d DisabledVcnImpl) ListNodes(_ context.Context) ([]domain.VcnNode, error) {
	return nil, rerrors.Wrap(user_errors.ErrServiceIsDisabled)
}
