package matreshka

import (
	errors "go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka_be_api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	AuthHeader = "Matreshka-Auth"

	ServiceName = "matreshka"
)

type Client interface {
	matreshka_be_api.MatreshkaBeAPIClient
}

func NewClient(opts ...grpc.DialOption) (Client, error) {
	opts = append(opts,
		// TODO Add token for matreshka
		//grpc.WithUnaryInterceptor(security.HeaderOutgoingInterceptor(AuthHeader, token)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	dial, err := grpc.NewClient("verv://"+ServiceName, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing")
	}

	return matreshka_be_api.NewMatreshkaBeAPIClient(dial), nil
}
