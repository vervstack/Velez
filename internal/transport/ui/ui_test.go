package ui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewServer_ServesRealFiles(t *testing.T) {
	handler := NewServer()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/velez.ico", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func Test_NewServer_ServesRealAssetFile(t *testing.T) {
	handler := NewServer()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/assets/index-BgRjN6ge.css", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func Test_NewServer_FallsBackToIndexForUnknownRoutes(t *testing.T) {
	indexReq := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	indexRec := httptest.NewRecorder()
	NewServer().ServeHTTP(indexRec, indexReq)
	require.Equal(t, http.StatusOK, indexRec.Code)

	for _, route := range []string{"/cp", "/vcn", "/smerd/some-name", "/deploy"} {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, route, nil)
		rec := httptest.NewRecorder()
		NewServer().ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code, "route %s should fall back to 200", route)
		require.Equal(t, indexRec.Body.String(), rec.Body.String(), "route %s should serve index.html body", route)
	}
}
