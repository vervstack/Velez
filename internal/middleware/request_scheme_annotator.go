package middleware

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

// RequestSchemeAnnotator is a runtime.WithMetadata hook (see transport.NewGatewayMux) that
// records whether the originating HTTP request was secure, so CookieForwardResponseOption can
// decide the Secure attribute on any cookie it sets without needing a static config value —
// grpc-gateway invokes every WithMetadata annotator with the raw *http.Request per call, and the
// returned metadata lands in the same context.Context later passed into
// runtime.WithForwardResponseOption hooks.
func RequestSchemeAnnotator(_ context.Context, r *http.Request) metadata.MD {
	if !requestIsSecure(r) {
		return nil
	}

	return metadata.Pairs(RequestSecureKey, RequestSecureValue)
}

// requestIsSecure detects the request's original scheme. X-Forwarded-Proto wins if present, and
// must equal "https" exactly (case-insensitive) to count as secure; any other value is treated as
// insecure rather than falling through to the r.TLS check. Absent the header, r.TLS != nil covers
// a direct TLS-terminated connection.
func requestIsSecure(r *http.Request) bool {
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return strings.EqualFold(proto, "https")
	}

	return r.TLS != nil
}
