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

func SetupServiceDiscovery(addr string, token string) (makosh_be.MakoshBeAPIClient, error) {
	_ = os.Setenv(makosh_resolver.MakoshURL, addr)
	_ = os.Setenv(makosh_resolver.MakoshSecret, token)

	_, err := vervResolver.Init()
	if err != nil {
		return nil, errors.Wrap(err, "error initializing verv-service-discovery")
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	apiClient, err := makosh.New(token, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error creating makosh api client")
	}

	_, err = apiClient.Version(context.Background(), &makosh_be.Version_Request{})
	if err != nil {
		return nil, errors.Wrap(err, "error pinging service discovery")
	}

	return apiClient, nil
}
