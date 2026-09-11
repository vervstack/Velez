package user_errors

import (
	"go.redsock.ru/rerrors"
)

// ErrVpnResultAlreadyExists is returned by vpnconnect.Runner.Result when the
// tailscale sidecar this run would have created is already up. Both call
// sites (internal/cluster/configuration, internal/cluster/service_discovery)
// branch on it and treat it as success rather than a failure.
var ErrVpnResultAlreadyExists = rerrors.New("pipeline result already exists")
