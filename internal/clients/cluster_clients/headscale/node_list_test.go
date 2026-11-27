package headscale

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/service/user_errors"
)

func Test_Client_ListNodes_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/v1/node", r.URL.Path)
		require.Equal(t, "Bearer "+testAPIKey, r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, writeErr := w.Write([]byte(`{
			"nodes": [
				{
					"id": "1",
					"name": "node-1",
					"givenName": "svc-ts-sidecar",
					"ipAddresses": ["100.64.0.1", "fd7a:115c:a1e0::1"],
					"online": true
				},
				{
					"id": 2,
					"name": "node-2-fallback",
					"givenName": "",
					"ipAddresses": ["100.64.0.2"]
				}
			]
		}`))
		require.NoError(t, writeErr)
	}))
	defer server.Close()

	client := &Client{
		headscaleApiUrl: server.URL,
		apiKey:          testAPIKey,
	}

	nodes, err := client.ListNodes(context.Background())
	require.NoError(t, err)
	require.Len(t, nodes, 2)

	require.Equal(t, "1", nodes[0].Id)
	require.Equal(t, "svc-ts-sidecar", nodes[0].Name)
	require.Equal(t, []string{"100.64.0.1", "fd7a:115c:a1e0::1"}, nodes[0].IpAddresses)
	require.True(t, nodes[0].IsOnline)

	require.Equal(t, "2", nodes[1].Id)
	require.Equal(t, "node-2-fallback", nodes[1].Name)
	require.Equal(t, []string{"100.64.0.2"}, nodes[1].IpAddresses)
	require.False(t, nodes[1].IsOnline)
}

func Test_Client_ListNodes_NonOkStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{
		headscaleApiUrl: server.URL,
		apiKey:          testAPIKey,
	}

	_, err := client.ListNodes(context.Background())
	require.ErrorIs(t, err, user_errors.ErrHeadscaleUnexpectedStatus)
}
