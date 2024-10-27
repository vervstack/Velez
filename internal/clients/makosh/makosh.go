package makosh

import (
	errors "github.com/Red-Sock/trace-errors"
	makosh "github.com/godverv/makosh/pkg/makosh_be"
	pb "github.com/godverv/makosh/pkg/makosh_be"
	"google.golang.org/grpc"

	"github.com/godverv/Velez/internal/clients/security"
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
