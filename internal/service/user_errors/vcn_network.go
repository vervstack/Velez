package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	ErrExternalNetworkRequiresCluster = rerrors.NewUserError(
		"external networks require cluster mode", codes.FailedPrecondition,
	)
	ErrVcnNetworkSourceRequired = rerrors.NewUserError("a network source is required", codes.InvalidArgument)
	ErrVcnInvalidHeadscaleUrl   = rerrors.NewUserError(
		"a valid HTTP or HTTPS Headscale URL is required", codes.InvalidArgument,
	)
	ErrVcnHeadscaleTokenRequired = rerrors.NewUserError("a Headscale API token is required", codes.InvalidArgument)
	ErrVcnNetworkNotFound        = rerrors.NewUserError("VPN network not found", codes.NotFound)
	ErrVcnNetworkExists          = rerrors.NewUserError(
		"a VPN network with that name already exists", codes.AlreadyExists,
	)
	ErrVcnInvalidPort = rerrors.NewUserError(
		"Headscale port must be between 1 and 65535", codes.InvalidArgument,
	)
	ErrVcnReachableHostRequired = rerrors.NewUserError(
		"reach this Velez node by an address other nodes can use to deploy Headscale", codes.InvalidArgument,
	)
	ErrVcnProviderUnsupported    = rerrors.NewUserError("unsupported VPN network provider", codes.FailedPrecondition)
	ErrVcnManagementKeyProtected = rerrors.NewUserError(
		"the Velez management API key cannot be expired", codes.FailedPrecondition,
	)
	ErrVcnKeyPrefixRequired    = rerrors.NewUserError("an API key prefix is required", codes.InvalidArgument)
	ErrVcnNamespaceIdMissing   = rerrors.New("Headscale returned a namespace without an ID")
	ErrVcnConnectionInProgress = rerrors.NewUserError(
		"a different VPN connection for this service is already in progress", codes.FailedPrecondition,
	)
	ErrVcnNetworkOwnerMismatch      = rerrors.New("VPN network belongs to a different node")
	ErrVcnHeadscaleNotRunning       = rerrors.New("Headscale network container is unavailable")
	ErrVcnNetworkServiceUnavailable = rerrors.New("VPN network service is unavailable")
)
