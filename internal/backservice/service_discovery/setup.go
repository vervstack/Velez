package service_discovery

import (
	"context"
	"os"

	errors "go.redsock.ru/rerrors"
	"go.vervstack.ru/makosh/pkg/makosh_be"
	vervResolver "go.vervstack.ru/makosh/pkg/resolver"
	"go.vervstack.ru/makosh/pkg/resolver/makosh_resolver"
	"go.vervstack.ru/matreshka/pkg/matreshka"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/clients/makosh"
)

type ServiceDiscovery struct {
	Sd *vervResolver.ServiceDiscovery
	makosh_be.MakoshBeAPIClient
}

func SetupServiceDiscovery(addr string, token string, msd matreshka.ServiceDiscovery) (sd ServiceDiscovery, err error) {
	_ = os.Setenv(makosh_resolver.MakoshURL, addr)
	_ = os.Setenv(makosh_resolver.MakoshSecret, token)

	sd.Sd, err = vervResolver.Init()
	if err != nil {
		return sd, errors.Wrap(err, "error initializing verv-service-discovery")
	}

	sd.Sd.SetOverrides(msd.Overrides)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	sd.MakoshBeAPIClient, err = makosh.New(token, opts...)
	if err != nil {
		return sd, errors.Wrap(err, "error creating makosh api client")
	}

	_, err = sd.Version(context.Background(), &makosh_be.Version_Request{})
	if err != nil {
		return sd, errors.Wrap(err, "error pinging service discovery")
	}

	return sd, nil
}

func (s *ServiceDiscovery) GetAddrs(vervName string) []string {
	matreshkaResolverPtr, err := s.Sd.GetResolver(vervName)
	if err != nil {
		return nil
	}

	return matreshkaResolverPtr.Load().GetAddrs()
}
