package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

// ErrServiceIsDisabled is returned by cluster_clients.ClusterClients' no-op
// implementations (internal/cluster/disabled.go,
// internal/cluster/verv_closed_network/no_impl.go) when cluster mode isn't
// enabled. Checked with rerrors.Is across internal/cluster/configuration,
// internal/cluster/service_discovery, internal/cluster/vpnconnect,
// internal/transport/velez_api_impl, internal/workers, and
// internal/service/service_manager/configurator to treat a disabled service
// as an expected outcome rather than a hard failure.
var ErrServiceIsDisabled = rerrors.New("service is disabled", codes.FailedPrecondition)
