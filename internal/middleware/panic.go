package middleware

import (
	"context"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc"
)

var (
	ErrPanicCough = rerrors.New("panic cough")
)

func PanicInterceptor() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			defer func() {
				if r := recover(); r != nil {
					e, ok := r.(error)
					if !ok {
						e = rerrors.Wrap(ErrPanicCough, r)
					}

					err = e
					log.Error().Err(err).Msg("panic in grpc handler")
				}
			}()

			return handler(ctx, req)
		})
}
