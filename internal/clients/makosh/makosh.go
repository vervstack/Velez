package makosh

import (
	errors "go.redsock.ru/rerrors"
	makosh "go.vervstack.ru/makosh/pkg/makosh_be"
	pb "go.vervstack.ru/makosh/pkg/makosh_be"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/middleware/security"
)

const (
	AuthHeader = "Makosh-Auth"

	ServiceName = "makosh"
)

func New(token string, opts ...grpc.DialOption) (makosh.MakoshBeAPIClient, error) {
	opts = append(opts,
		grpc.WithUnaryInterceptor(security.HeaderOutgoingInterceptor(AuthHeader, token)))

	dial, err := grpc.NewClient("verv://"+ServiceName, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing")
	}

	return pb.NewMakoshBeAPIClient(dial), nil
}
