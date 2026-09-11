package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrNotFound is returned by headscale.Client for a 404/empty result from
	// the Headscale API. Checked with rerrors.Is by
	// internal/cluster/vpnconnect/steps.go and
	// internal/jobs/connect_service_to_vpn.go to treat a missing preauth key as
	// expected rather than a hard failure.
	ErrNotFound = rerrors.New("not found")

	// ErrHeadscaleUnexpectedStatus is returned by headscale.Client when the
	// Headscale API responds with a status code none of its callers handle
	// explicitly.
	ErrHeadscaleUnexpectedStatus = rerrors.New("unexpected status")

	// ErrHeadscaleCantParseOutput is returned when the output of a headscale
	// CLI invocation (exec'd inside the headscale container) doesn't match
	// the format the caller expects.
	ErrHeadscaleCantParseOutput = rerrors.New("can't parse output")

	// ErrHeadscaleNamespaceAlreadyExists is returned by
	// headscale.Client.CreateNamespace when the requested namespace already
	// exists - a purely user-facing outcome, not a wiring bug.
	ErrHeadscaleNamespaceAlreadyExists = rerrors.NewUserError("namespace already exists", codes.AlreadyExists)

	// ErrHeadscaleNotConnectedToVervnet is returned when the headscale
	// container is up but isn't attached to the vervnet Docker network.
	ErrHeadscaleNotConnectedToVervnet = rerrors.New("headscale container isn't connected to vervnet")

	// ErrHeadscaleNoAliases is returned when the headscale container has no
	// network aliases to resolve it by.
	ErrHeadscaleNoAliases = rerrors.New("headscale container doesn't have any aliases")
)
