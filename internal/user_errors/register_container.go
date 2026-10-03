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

	// ErrContainerIsVelez is returned by RegisterContainer for a container
	// running a Velez image.
	ErrContainerIsVelez = rerrors.New("Velez containers can't be registered", codes.FailedPrecondition)

	// ErrContainerSharesNetwork is returned by RegisterContainer for a
	// container that runs in another container's network namespace.
	ErrContainerSharesNetwork = rerrors.New(
		"container shares another container's network, register the network root instead",
		codes.FailedPrecondition,
	)

	// ErrSidecarLeftoverExists is returned by recreate_sidecars when a
	// previous attempt left a "<name>_old" container next to a live sidecar.
	ErrSidecarLeftoverExists = rerrors.New("sidecar backup container already exists", codes.FailedPrecondition)

	// ErrContainerBindingsUnavailable is returned by register_container when
	// the live storage backend exposes no container bindings.
	ErrContainerBindingsUnavailable = rerrors.New("container bindings storage is unavailable", codes.FailedPrecondition)

	// ErrRegisterContainerIdRequired is returned by RegisterContainer for an
	// empty container id.
	ErrRegisterContainerIdRequired = rerrors.New("container id is required", codes.InvalidArgument)

	// ErrRegisterServiceNameRequired is returned by RegisterContainer for an
	// empty service name.
	ErrRegisterServiceNameRequired = rerrors.New("service name is required", codes.InvalidArgument)

	// ErrBindMountNotLinked is returned by register_container's
	// inspect_container when the container has a bind mount that is not
	// listed in the request's bind mount links. The wrap message lists the
	// offending source -> target pairs.
	ErrBindMountNotLinked = rerrors.New("bind mounts are not linked", codes.FailedPrecondition)

	// ErrBindMountLinkUnknown is returned by register_container's
	// inspect_container for a bind mount link whose source is not a bind
	// mount of the container.
	ErrBindMountLinkUnknown = rerrors.New("bind mount link matches no bind mount of the container", codes.InvalidArgument)

	// ErrRegisterPortsWithKeepMapping is returned by RegisterContainer when
	// ports are given together with keep_port_mapping.
	ErrRegisterPortsWithKeepMapping = rerrors.New(
		"ports can't be set together with keep_port_mapping", codes.InvalidArgument)

	// ErrVolumeOptionsConflict is returned when a volume to create already
	// exists with a different driver or options.
	ErrVolumeOptionsConflict = rerrors.New("volume already exists with different options", codes.AlreadyExists)

	// ErrPgCredentialsRequired is returned by register_container's
	// verify_pg_login when neither the container env nor the request carries
	// the Postgres superuser and password.
	ErrPgCredentialsRequired = rerrors.New(
		"postgres superuser and password are required: the container env has none", codes.InvalidArgument)

	// ErrPgLoginFailed is returned by register_container's verify_pg_login
	// when a real login with the resolved credentials fails. It wraps the
	// driver error.
	ErrPgLoginFailed = rerrors.New("postgres login with the given credentials failed", codes.FailedPrecondition)

	// ErrRunnerTargetRequired is returned by RegisterContainer for a runner
	// pattern without a target.
	ErrRunnerTargetRequired = rerrors.New("runner target is required", codes.InvalidArgument)

	// ErrRegistryCredentialsRequired is returned by RegisterContainer for a
	// registry pattern carrying only one of username and password, and by
	// register_container's verify_registry_login when the container is
	// configured with htpasswd auth but the request carries no login.
	ErrRegistryCredentialsRequired = rerrors.New(
		"registry username and password are required: the container is configured with htpasswd auth",
		codes.InvalidArgument)

	// ErrRegistryLoginFailed is returned by register_container's
	// verify_registry_login when a real login against the registry with the
	// given credentials does not succeed.
	ErrRegistryLoginFailed = rerrors.New("registry login with the given credentials failed", codes.FailedPrecondition)

	// ErrNoRegistryPortExposure is returned by register_container's
	// verify_registry_login when Velez runs outside a container and the
	// registry container publishes no host port to log in through.
	ErrNoRegistryPortExposure = rerrors.New(
		"registry container publishes no host port to verify the login through", codes.FailedPrecondition)
)
