package user_errors

import (
	"go.redsock.ru/rerrors"
)

var (
	// ErrStorageAlreadyExists is returned by the internal/storage.* backends
	// (registries, environments, secrets, pg_instances, postgres) when a
	// create collides with an existing row.
	ErrStorageAlreadyExists = rerrors.New("already exists")

	// ErrStorageNotFound is returned by the internal/storage.* backends for a
	// missing row, wrapped with the specific lookup that failed.
	ErrStorageNotFound = rerrors.New("not found")
)
