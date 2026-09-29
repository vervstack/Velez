package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrRegisterContainerNotFound is returned by register_container's
	// inspect_container when the container id resolves to no container.
	ErrRegisterContainerNotFound = rerrors.New("container to register not found", codes.NotFound)

	// ErrContainerAlreadyLinked is returned by register_container's
	// inspect_container when the container already carries a service label
	// or has a binding.
	ErrContainerAlreadyLinked = rerrors.New("container is already linked to a service", codes.AlreadyExists)

	// ErrContainerIsSidecar is returned by register_container's
	// inspect_container for a Velez sidecar container.
	ErrContainerIsSidecar = rerrors.New("sidecar containers can't be registered", codes.FailedPrecondition)

	// ErrContainerBindingsUnavailable is returned by register_container when
	// the live storage backend exposes no container bindings.
	ErrContainerBindingsUnavailable = rerrors.New("container bindings storage is unavailable", codes.FailedPrecondition)

	// ErrRegisterContainerIdRequired is returned by RegisterContainer for an
	// empty container id.
	ErrRegisterContainerIdRequired = rerrors.New("container id is required", codes.InvalidArgument)

	// ErrRegisterServiceNameRequired is returned by RegisterContainer for an
	// empty service name.
	ErrRegisterServiceNameRequired = rerrors.New("service name is required", codes.InvalidArgument)
)
