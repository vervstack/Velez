package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrRegistryUnexpectedStatus is returned by registryclients.Client when
	// a registry's HTTP API responds with a status code none of its callers
	// handle explicitly.
	ErrRegistryUnexpectedStatus = rerrors.New("unexpected status")

	// ErrRegistryNameRequired is returned when CreateRegistry/UpdateRegistry
	// carries an empty name.
	ErrRegistryNameRequired = rerrors.New("registry name is required", codes.InvalidArgument)

	// ErrRegistryNotFound is returned when a registry id/name a caller passed
	// doesn't resolve to a row in velez.registries.
	ErrRegistryNotFound = rerrors.New("registry not found", codes.NotFound)

	// ErrInvalidRegistryType is returned when Type isn't one of
	// domain.RegistryTypeDockerHub / domain.RegistryTypeGenericV2.
	ErrInvalidRegistryType = rerrors.New("invalid registry type", codes.InvalidArgument)

	// ErrRegistryIdRequired is returned when UpdateRegistry carries no ID.
	ErrRegistryIdRequired = rerrors.New("registry id is required", codes.InvalidArgument)

	// ErrRegistryIdOrNameRequired is returned when DeleteRegistry carries
	// neither an ID nor a name to resolve the target by.
	ErrRegistryIdOrNameRequired = rerrors.New("registry id or name is required", codes.InvalidArgument)
)
