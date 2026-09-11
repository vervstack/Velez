package user_errors

import (
	"go.redsock.ru/rerrors"
)

// ErrNotFound is returned by headscale.Client for a 404/empty result from
// the Headscale API. Checked with rerrors.Is by
// internal/cluster/vpnconnect/steps.go and
// internal/jobs/connect_service_to_vpn.go to treat a missing preauth key as
// expected rather than a hard failure.
var ErrNotFound = rerrors.New("not found")
