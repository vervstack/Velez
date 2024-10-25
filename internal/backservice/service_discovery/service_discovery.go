package service_discovery

import (
	"context"
	"os"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/makosh/pkg/makosh_be"
	vervResolver "github.com/godverv/makosh/pkg/resolver"
	"github.com/godverv/makosh/pkg/resolver/makosh_resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/clients/makosh"
)

type ServiceDiscovery struct {
	Api     makosh_be.MakoshBeAPIClient
	LocalSD *vervResolver.ServiceDiscovery
}

func NewServiceDiscovery(addr string, token string) (*ServiceDiscovery, error) {
	//remoteResolver, err := makosh_resolver.NewBuilder(
	//	makosh_resolver.WithURL(addrs),
	//	makosh_resolver.WithSecret(token),
	//)
	//
	//serviceDiscovery, err := vervResolver.NewLocalServiceDiscovery(
	//	vervResolver.WithResolverBuilder(remoteResolver),
	//)
	//if err != nil {
	//	return nil, errors.Wrap(err, "error creating makosh resolver")
	//}
	//
	//grpcResolver.Register(serviceDiscovery.GrpcBuilder())

	_ = os.Setenv(makosh_resolver.MakoshURL, addr)
	_ = os.Setenv(makosh_resolver.MakoshSecret, token)

	serviceDiscovery, err := vervResolver.Init()
	if err != nil {
		return nil, errors.Wrap(err, "error initializing verv-service-discovery")
	}

	//rPtr, err := serviceDiscovery.GetResolver(makosh.ServiceName)
	//if err != nil {
	//	return nil, errors.Wrap(err, "error getting resolver")
	//}

	//err = rPtr.Load().SetAddrs(addr)
	//if err != nil {
	//	return nil, errors.Wrap(err, "error manually setting addresses")
	//}

	apiClient, err := makosh.New(token,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating makosh api client")
	}

	_, err = apiClient.Version(context.Background(), &makosh_be.Version_Request{})
	if err != nil {
		return nil, errors.Wrap(err, "error pinging service discovery")
	}

	return &ServiceDiscovery{
		Api:     apiClient,
		LocalSD: serviceDiscovery,
	}, nil
}
