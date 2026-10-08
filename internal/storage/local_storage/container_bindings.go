package local_storage

import (
	"context"
	"database/sql"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
)

type containerBindings struct{}

func newContainerBindingsStorage() *containerBindings {
	return &containerBindings{}
}

func (c *containerBindings) Upsert(_ context.Context, _ domain.ContainerBinding) error {
	return rerrors.Wrap(errBindingsUnsupported)
}

// Delete is a no-op: single-node mode never stores a binding (see Upsert).
func (c *containerBindings) Delete(_ context.Context, _ int32, _, _ string) error {
	return nil
}

func (c *containerBindings) ListByNode(_ context.Context, _ int32, _ string) ([]domain.ContainerBinding, error) {
	return []domain.ContainerBinding{}, nil
}

func (c *containerBindings) WithTx(_ *sql.Tx) storage.ContainerBindingsStorage {
	return c
}
