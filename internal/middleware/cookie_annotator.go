package middleware

import (
	"context"
	"net/http"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/metadata"
)

// authHeader is the gRPC metadata key carrying the bearer credential, matching whatever a
// project's own auth interceptor reads off incoming metadata.
const (
	authHeader = "authorization"
)

// CookieToMetadataAnnotator is registered via runtime.WithMetadata on the shared gateway mux
// (see transport.NewGatewayMux). It translates the access_token cookie into the "authorization"
// gRPC metadata key that a project's own auth interceptor reads, so browser callers
// authenticate the same way native gRPC/CLI callers do.
//
// It only fires when the request carries neither an "Authorization" nor a
// "Grpc-Metadata-Authorization" header — an explicit header always wins over the cookie,
// deterministically, rather than relying on grpc-gateway's metadata.Join merge order.
func CookieToMetadataAnnotator(_ context.Context, r *http.Request) metadata.MD {
	if r.Header.Get("Authorization") != "" || r.Header.Get("Grpc-Metadata-Authorization") != "" {
		return nil
	}

	cookie, err := r.Cookie(AccessTokenCookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}

	return metadata.Pairs(
		authHeader, cookie.Value,
		AuthViaCookieMarkerKey, AuthViaCookieMarkerValue,
	)
}

// CookieValueFromRawHeader extracts a single cookie's value out of a raw "Cookie" header
// string (as carried in the GatewayCookieMetadataKey gRPC metadata value, rather than a live
// *http.Request). Used by GrpcCSRFInterceptor and any Refresh-style handler that needs to read
// a cookie from gRPC metadata rather than an http.Request.
func CookieValueFromRawHeader(rawCookieHeader, name string) (string, error) {
	header := http.Header{}
	header.Add("Cookie", rawCookieHeader)

	req := http.Request{Header: header}

	cookie, err := req.Cookie(name)
	if err != nil {
		return "", rerrors.Wrap(err, "error reading cookie from raw header")
	}

	return cookie.Value, nil
}
