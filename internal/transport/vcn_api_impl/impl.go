package vcn_api_impl

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/pipelines"
	"google.golang.org/grpc"
)

type Impl struct {
	velez_api.UnimplementedVcnApiServer

	vpnService cluster_clients.VervClosedNetworkClient
	pipeliner  pipelines.Pipeliner
	jobsEngine jobs.Engine
}

func New(cluster cluster_clients.ClusterClients, pipeliner pipelines.Pipeliner, jobsEngine jobs.Engine) *Impl {
	return &Impl{
		vpnService: cluster.Vpn(),
		pipeliner:  pipeliner,
		jobsEngine: jobsEngine,
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	velez_api.RegisterVcnApiServer(server, impl)
}

func (impl *Impl) Gateway(
	ctx context.Context,
	endpoint string,
	opts ...grpc.DialOption,
) (route string, handler http.Handler) {
	gwHTTPMux := runtime.NewServeMux()

	err := velez_api.RegisterVcnApiHandlerFromEndpoint(
		ctx,
		gwHTTPMux,
		endpoint,
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/vcn/", gwHTTPMux
}
