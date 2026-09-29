package local_storage

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

type containerBindings struct{}

func newContainerBindingsStorage() *containerBindings {
	return &containerBindings{}
}

func (c *containerBindings) Upsert(_ context.Context, _ domain.ContainerBinding) error {
	return rerrors.Wrap(errBindingsUnsupported)
}

func (c *containerBindings) ListByNode(_ context.Context, _ int32, _ string) ([]domain.ContainerBinding, error) {
	return []domain.ContainerBinding{}, nil
}
