package matreshka

import (
	errors "go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	AuthHeader = "Matreshka-Auth"

	ServiceName = "matreshka"
)

type Client interface {
	matreshka_api.MatreshkaBeAPIClient
}

func NewClient(opts ...grpc.DialOption) (Client, error) {
	opts = append(opts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	dial, err := grpc.NewClient("verv://"+ServiceName, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing")
	}

	return matreshka_api.NewMatreshkaBeAPIClient(dial), nil
}
