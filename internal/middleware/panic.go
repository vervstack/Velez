package middleware

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func PanicInterceptor() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			defer func() {
				if r := recover(); r != nil {
					e, ok := r.(error)
					if !ok {
						e = fmt.Errorf("panic: %v", r)
					}

					err = e
					log.Error().Err(err).Msg("panic in grpc handler")
				}
			}()

			return handler(ctx, req)
		})
}
