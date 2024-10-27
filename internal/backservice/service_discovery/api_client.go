package service_discovery

import (
	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/makosh/pkg/makosh_be"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/clients/makosh"
	"github.com/godverv/Velez/internal/clients/security"
)

type ApiClient struct {
	makosh_be.MakoshBeAPIClient
}

func newApiClient(addr string, token string) (*ApiClient, error) {
	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(
			security.HeaderOutgoingInterceptor(makosh.AuthHeader, token)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	dial, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing")
	}

	client := &ApiClient{
		MakoshBeAPIClient: makosh_be.NewMakoshBeAPIClient(dial),
	}

	return client, nil
}
