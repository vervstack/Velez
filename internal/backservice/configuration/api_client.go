package configuration

import (
	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka-be/pkg/matreshka_be_api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ApiClient struct {
	matreshka_be_api.MatreshkaBeAPIClient
}

func newApiClient(addr string) (*ApiClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	dial, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing")
	}

	client := &ApiClient{
		MatreshkaBeAPIClient: matreshka_be_api.NewMatreshkaBeAPIClient(dial),
	}

	return client, nil
}
