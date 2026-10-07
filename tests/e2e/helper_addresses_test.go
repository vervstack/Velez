//go:build e2e_full

package e2e

import (
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

const (
	addressesRebuildTimeout      = 60 * time.Second
	addressesRebuildPollInterval = 200 * time.Millisecond
)

func singleNodePlane() Plane {
	return Planes[0]
}

func newRebuildAddressesRequest() *velez_api.RebuildAddresses_Request {
	return &velez_api.RebuildAddresses_Request{}
}

func rebuildAddresses(t *testing.T, env *TestEnvironment) {
	t.Helper()

	req := newRebuildAddressesRequest()

	_, err := env.Custom.SettingsApiImpl.RebuildAddresses(t.Context(), req)
	require.NoError(t, err)

	status := waitForAddressesRebuildIdle(t, env)
	require.Empty(t, status.GetLastError())
}

func newGetAddressesRebuildStatusRequest() *velez_api.GetAddressesRebuildStatus_Request {
	return &velez_api.GetAddressesRebuildStatus_Request{}
}

func waitForAddressesRebuildIdle(t *testing.T, env *TestEnvironment) *velez_api.GetAddressesRebuildStatus_Response {
	t.Helper()

	req := newGetAddressesRebuildStatusRequest()

	var status *velez_api.GetAddressesRebuildStatus_Response

	require.Eventually(t, func() bool {
		resp, err := env.Custom.SettingsApiImpl.GetAddressesRebuildStatus(t.Context(), req)
		if err != nil {
			return false
		}

		status = resp

		return !resp.GetIsRunning()
	}, addressesRebuildTimeout, addressesRebuildPollInterval)

	return status
}

func expectedDockerAddressHost(t *testing.T) string {
	t.Helper()

	parsed, err := url.Parse(os.Getenv("DOCKER_HOST"))
	require.NoError(t, err)

	return parsed.Hostname()
}

func requireSingleDockerAddress(
	t *testing.T, resource *velez_api.BoundResource, wantPort uint32,
) *velez_api.ResourceAddress {
	t.Helper()

	addresses := resource.GetAddresses()
	require.Len(t, addresses, 1, "resource addresses: %v", addresses)

	address := addresses[0]
	require.Equal(t, velez_api.AddressScope_ADDRESS_SCOPE_DOCKER, address.GetScope())
	require.Equal(t, wantPort, address.GetPort())
	require.Equal(t, expectedDockerAddressHost(t), address.GetHost())

	return address
}
