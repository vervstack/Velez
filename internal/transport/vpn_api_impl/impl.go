package vpn_api_impl

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/cluster"
	"go.vervstack.ru/Velez/internal/pipelines"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type Impl struct {
	velez_api.UnimplementedVpnApiServer

	vpnService cluster_clients.VervPrivateNetworkClient
	pipeliner  pipelines.Pipeliner
}

func New(cluster cluster.Cluster, pipeliner pipelines.Pipeliner) *Impl {
	return &Impl{
		vpnService: cluster.Vpn(),
		pipeliner:  pipeliner,
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	velez_api.RegisterVpnApiServer(server, impl)
}

func (impl *Impl) Gateway(ctx context.Context, endpoint string, opts ...grpc.DialOption) (route string, handler http.Handler) {
	gwHttpMux := runtime.NewServeMux()

	err := velez_api.RegisterVpnApiHandlerFromEndpoint(
		ctx,
		gwHttpMux,
		endpoint,
		opts,
	)
	if err != nil {
		logrus.Errorf("error registering grpc2http handler: %s", err)
	}

	return "/api/vpn/", gwHttpMux
}
