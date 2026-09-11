package middleware

import (
	"context"
	"crypto/subtle"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// csrfHeaderMetadataKey is the gRPC metadata key for the frontend's CSRF echo, sent as the
// Grpc-Metadata-X-Csrf-Token HTTP header (grpc-gateway's DefaultHeaderMatcher strips the
// Grpc-Metadata- prefix and lowercases the rest).
const (
	csrfHeaderMetadataKey = "x-csrf-token"
)

// GrpcCSRFInterceptor enforces double-submit CSRF protection, deny-by-default, for RPCs not in
// exemptMethods. It only fires for requests CookieToMetadataAnnotator authenticated from a
// cookie (marked via AuthViaCookieMarkerKey); native gRPC/CLI callers presenting a raw
// Authorization header never carry that marker and stay exempt.
func GrpcCSRFInterceptor(exemptMethods ...string) grpc.ServerOption {
	interceptor := NewCSRFInterceptor(exemptMethods...)

	return grpc.ChainUnaryInterceptor(interceptor)
}

// NewCSRFInterceptor builds the interceptor function itself, kept separate from
// GrpcCSRFInterceptor so it can be exercised directly in unit tests without spinning up a real
// grpc.Server.
func NewCSRFInterceptor(exemptMethods ...string) grpc.UnaryServerInterceptor {
	exempt := make(map[string]struct{}, len(exemptMethods))
	for _, m := range exemptMethods {
		exempt[m] = struct{}{}
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		if !authenticatedViaCookie(md) {
			return handler(ctx, req)
		}

		_, isExempt := exempt[info.FullMethod]
		if isExempt {
			return handler(ctx, req)
		}

		if !csrfTokenValid(md) {
			return nil, status.Error(codes.PermissionDenied, "csrf token missing or invalid")
		}

		return handler(ctx, req)
	}
}

func authenticatedViaCookie(md metadata.MD) bool {
	return metadataValue(md, AuthViaCookieMarkerKey) == AuthViaCookieMarkerValue
}

func csrfTokenValid(md metadata.MD) bool {
	rawCookieHeader := metadataValue(md, GatewayCookieMetadataKey)
	if rawCookieHeader == "" {
		return false
	}

	cookieValue, err := CookieValueFromRawHeader(rawCookieHeader, CSRFCookieName)
	if err != nil || cookieValue == "" {
		return false
	}

	headerValue := metadataValue(md, csrfHeaderMetadataKey)
	if headerValue == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(cookieValue), []byte(headerValue)) == 1
}
