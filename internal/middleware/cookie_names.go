package middleware

// Cookie names set on the browser by CookieForwardResponseOption and read back by
// CookieToMetadataAnnotator / GrpcCSRFInterceptor / auth handlers' Refresh fallback.
const (
	// AccessTokenCookieName carries the same raw token value as the "authorization" gRPC
	// metadata key — no Bearer prefix, no transformation.
	AccessTokenCookieName = "access_token"
	// RefreshTokenCookieName is read by an auth handler's Refresh handler when the request
	// body's refresh_token field is empty (the browser path never sends it in the body).
	RefreshTokenCookieName = "refresh_token"
	// CSRFCookieName is the non-httpOnly double-submit cookie; browsers echo its value back
	// via the Grpc-Metadata-X-Csrf-Token header, compared in GrpcCSRFInterceptor.
	CSRFCookieName = "csrf_token"
	// CookiePath scopes the httpOnly session cookies to the gateway's API root — adjust to
	// match wherever this project mounts its gRPC-gateway routes.
	CookiePath = "/api/"
	// CSRFCookiePath is deliberately "/" rather than CookiePath: a cookie's Path attribute
	// gates document.cookie visibility exactly like it gates which requests carry the cookie,
	// and the frontend typically needs to read the csrf_token from any app route, not just
	// API routes, to echo it back.
	CSRFCookiePath = "/"

	// SetCookieAccessTokenKey and the keys below are metadata keys used between auth handlers
	// and CookieForwardResponseOption to signal which cookies to set/clear on the HTTP
	// response. Set via grpc.SetHeader in the handler, read back via
	// runtime.ServerMetadataFromContext in the forward-response hook.
	SetCookieAccessTokenKey        = "x-set-cookie-access-token"
	SetCookieAccessTokenExpiryKey  = "x-set-cookie-access-token-expiry"
	SetCookieRefreshTokenKey       = "x-set-cookie-refresh-token"
	SetCookieRefreshTokenExpiryKey = "x-set-cookie-refresh-token-expiry"
	ClearAuthCookiesKey            = "x-clear-auth-cookies"

	// ClearAuthCookiesValue is the fixed value paired with ClearAuthCookiesKey — a boolean flag,
	// not a token, so a fixed sentinel is enough.
	ClearAuthCookiesValue = "1"

	// RequestSecureKey is the gRPC metadata key RequestSchemeAnnotator sets whenever it
	// determined the originating HTTP request was secure (HTTPS, directly or via a trusted
	// X-Forwarded-Proto). Absence of the key means false — there is no "0" sentinel value.
	RequestSecureKey = "x-request-secure"
	// RequestSecureValue is the fixed value paired with RequestSecureKey.
	RequestSecureValue = "1"

	// AuthViaCookieMarkerKey is the gRPC metadata key CookieToMetadataAnnotator sets whenever it
	// supplied the "authorization" credential from a cookie rather than a header.
	// GrpcCSRFInterceptor only enforces CSRF checks when this marker is present, so native
	// gRPC/CLI callers presenting a raw Authorization header stay exempt.
	AuthViaCookieMarkerKey = "x-auth-via-cookie"
	// AuthViaCookieMarkerValue is the fixed value paired with AuthViaCookieMarkerKey.
	AuthViaCookieMarkerValue = "1"

	// GatewayCookieMetadataKey is the gRPC metadata key grpc-gateway's DefaultHeaderMatcher
	// populates from the HTTP request's raw "Cookie" header (a permanent IANA header, forwarded
	// with the "grpcgateway-" prefix — see grpc-gateway/v2/runtime.DefaultHeaderMatcher).
	GatewayCookieMetadataKey = "grpcgateway-cookie"
)
