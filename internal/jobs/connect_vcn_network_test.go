package jobs

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

type testNetworkConnection func(context.Context, string) (string, string, error)

func (f testNetworkConnection) Connection(ctx context.Context, id string) (string, string, error) {
	return f(ctx, id)
}

func TestResolveVcnNetworkJob(t *testing.T) {
	t.Parallel()

	const (
		networkId = "network-id"
		nodeName  = "node-name"
		loginUrl  = "https://headscale.example"
	)

	input := &pb.ConnectServiceToVpnTaskPayload{NetworkId: proto(networkId), NodeName: nodeName}
	encoded, err := json.Marshal(input)
	require.NoError(t, err)

	var payload pb.ConnectServiceToVpnTaskPayload

	err = json.Unmarshal(encoded, &payload)
	require.NoError(t, err)
	require.Equal(t, nodeName, payload.GetNodeName())

	connector := testNetworkConnection(func(_ context.Context, id string) (string, string, error) {
		require.Equal(t, networkId, id)

		return loginUrl, "reusable-key", nil
	})

	handler := &connectServiceToVpnHandler{networks: connector}
	steps := handler.connectionJobs(&payload)
	require.Len(t, steps, 1)
	require.Equal(t, stepResolveVcnNetwork, steps[0].Name)

	err = steps[0].Job.Do(t.Context())
	require.NoError(t, err)
	require.Equal(t, loginUrl, payload.GetLoginServerUrl())
	require.Equal(t, "reusable-key", payload.GetClientKey())
}
