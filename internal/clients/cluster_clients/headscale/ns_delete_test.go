package headscale

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	testAPIKey = "test-key"
)

func Test_Client_DeleteNamespace_DeletesNodesFirst(t *testing.T) {
	deleted := make([]string, 0, 3)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer "+testAPIKey, r.Header.Get("Authorization"))

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/node":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			_, writeErr := w.Write([]byte(`{
				"nodes": [
					{"id": "1", "user": {"id": "7"}},
					{"id": 2, "user": {"id": "8"}},
					{"id": "3", "user": {"id": 7}}
				]
			}`))
			require.NoError(t, writeErr)
		case r.Method == http.MethodDelete:
			deleted = append(deleted, r.URL.Path)

			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := &Client{
		headscaleApiUrl: server.URL,
		apiKey:          testAPIKey,
	}

	err := client.DeleteNamespace(context.Background(), "7")
	require.NoError(t, err)

	require.Equal(t, []string{"/api/v1/node/1", "/api/v1/node/3", "/api/v1/user/7"}, deleted)
}

func Test_Client_DeleteNamespace_NodeDeleteFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			_, writeErr := w.Write([]byte(`{"nodes": [{"id": "1", "user": {"id": "7"}}]}`))
			require.NoError(t, writeErr)

			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{
		headscaleApiUrl: server.URL,
		apiKey:          testAPIKey,
	}

	err := client.DeleteNamespace(context.Background(), "7")
	require.ErrorIs(t, err, user_errors.ErrHeadscaleUnexpectedStatus)
}
