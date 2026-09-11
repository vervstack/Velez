package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrContainerIdMissing is returned by internal/jobs and
	// internal/cluster/vpnconnect steps that operate on a container but
	// received no container id in their TaskContext/step input.
	ErrContainerIdMissing = rerrors.New("no container id provided")

	// ErrContainerIdEmpty is returned by internal/jobs steps that operate on
	// a container id string that was present but empty.
	ErrContainerIdEmpty = rerrors.New("empty container id")

	// ErrContainerIdRequired is returned by internal/jobs' upgrade_smerd
	// steps that require a non-empty container id to proceed.
	ErrContainerIdRequired = rerrors.New("container id is required")

	// ErrContainerNotCreated is returned by internal/jobs' create_smerd
	// steps when a step that expects the container to already exist finds
	// it wasn't created.
	ErrContainerNotCreated = rerrors.New("container was not created")

	// ErrHealthcheckRetriesExhausted is returned by internal/jobs'
	// create_smerd steps when a container's healthcheck never turns
	// healthy within the retry budget.
	ErrHealthcheckRetriesExhausted = rerrors.New("healthcheck retries exhausted")

	// ErrSidecarAlreadyRunning is returned by internal/jobs'
	// connect_service_to_vpn steps when the VPN sidecar container is
	// already running.
	ErrSidecarAlreadyRunning = rerrors.New("sidecar container already running")

	// ErrInvalidServiceName is returned when CreateService is called with a
	// service name containing characters outside the allowed set.
	ErrInvalidServiceName = rerrors.New("service name contains invalid characters", codes.InvalidArgument)

	// ErrTooShortServiceName is returned when CreateService is called with a
	// service name shorter than the minimum length.
	ErrTooShortServiceName = rerrors.New("service name is too short", codes.InvalidArgument)

	// ErrRegistriesStorageMissingBuiltinUpsert signals a storage wiring bug,
	// not a user-facing condition - registries.NewStatic and
	// registries.NewPg both implement BuiltinRegistryUpserter, so this only
	// fires if a third RegistriesStorage implementation is ever wired in
	// without it.
	ErrRegistriesStorageMissingBuiltinUpsert = rerrors.New("registries storage does not support builtin upsert")

	// ErrPgContainerNotHealthy is returned by internal/jobs' enable_statefull
	// steps when the postgres container never becomes healthy within the
	// wait budget.
	ErrPgContainerNotHealthy = rerrors.New("timed out waiting for postgres container to become healthy")

	// ErrNoNetworkSettings is returned by internal/jobs' enable_statefull
	// steps when a container's inspect result carries no network settings.
	ErrNoNetworkSettings = rerrors.New("no network settings found in container")

	// ErrNoPgPortExposure is returned by internal/jobs' enable_statefull
	// steps when a postgres container exposes no mapping for port 5432.
	ErrNoPgPortExposure = rerrors.New("no exposure for 5432 found")

	// ErrSelfUpgradeIsForbidden is returned by internal/jobs' upgrade_smerd
	// steps when a request would upgrade Velez's own container - moved here
	// from the deleted internal/pipelines/steps/upgrade_steps package, where
	// it was the last symbol anything outside internal/pipelines still
	// imported.
	ErrSelfUpgradeIsForbidden = rerrors.NewUserError("Can't perform self upgrade", codes.FailedPrecondition)

	// ErrTaskConfigNil is returned by
	// internal/cluster/env/container_service_task when constructing a
	// TaskV2 with a nil container.CreateRequest.
	ErrTaskConfigNil = rerrors.New("config is nil")

	// ErrTaskHostnameEmpty is returned by
	// internal/cluster/env/container_service_task when constructing a
	// TaskV2 whose config carries an empty hostname.
	ErrTaskHostnameEmpty = rerrors.New("hostname is empty")

	// ErrJobFailedToExecute is returned by internal/jobs' checkpointedJob
	// when a job's checkpoint row was already left FAILED by a previous run
	// - the stored failure message is attached via rerrors.Wrap at the call
	// site, never rebuilt as an ad-hoc error.
	ErrJobFailedToExecute = rerrors.New("job previously failed")

	// ErrTaskFailed is returned across transport/worker call sites once a
	// task's FinishTask status settles on FAILED - the stored
	// finalTask.Error.String is attached via rerrors.Wrap, never rebuilt as
	// an ad-hoc error.
	ErrTaskFailed = rerrors.New("task failed")

	// ErrNoHandlerRegisteredForAction is returned by taskWorker.run when a
	// claimed task's Action has no TaskHandler registered for it.
	ErrNoHandlerRegisteredForAction = rerrors.New("no handler registered for action")

	// ErrContainerNotFoundInEnvironment is returned by internal/jobs'
	// upgrade_smerd steps when a lookup by name finds no container in the
	// target environment.
	ErrContainerNotFoundInEnvironment = rerrors.New("container not found in environment")

	// ErrPortAlreadyOccupied is returned by internal/jobs' enable_statefull
	// steps when the requested Postgres host port is already bound on this
	// node.
	ErrPortAlreadyOccupied = rerrors.New("requested port is already occupied on this node")
)
