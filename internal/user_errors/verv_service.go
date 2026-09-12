package user_errors

import (
	"go.redsock.ru/rerrors"
)

var (
	// ErrServiceNameRequiredToFind is returned when GetServiceReq carries no
	// name to look a service up by.
	ErrServiceNameRequiredToFind = rerrors.New("name is required to find service")

	// ErrNoRunningDeployment is returned when a service has no running
	// deployment to act on.
	ErrNoRunningDeployment = rerrors.New("no running deployment found for service")

	// ErrEmptyEnvironmentSuffix is returned when cascade-deleting an
	// environment's Docker resources is attempted with an empty suffix -
	// refused because an empty suffix would match resources across every
	// environment.
	ErrEmptyEnvironmentSuffix = rerrors.New(
		"refusing to cascade-delete resources for an environment with an empty suffix",
	)

	// ErrServiceHasRunningInstances is returned by RemoveService when the
	// service still has running instances and the caller didn't request
	// drop_running_instances.
	ErrServiceHasRunningInstances = rerrors.New(
		"service has running instances, stop/drop them first or set drop_running_instances",
	)
)
