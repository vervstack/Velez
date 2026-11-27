//go:build e2e_full

package e2e

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/patterns"
	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	vcnNetworksSuffix          = "e2e-vcn-networks"
	vcnNetworkLifecycleTimeout = 3 * time.Minute
	vcnNetworkRequestTimeout   = 10 * time.Second
)

func Test_VcnNetworks(t *testing.T) {
	t.Parallel()

	hs := getSharedHeadscale(t)
	env, _ := newStatefullEnvironment(t, Planes[1], vcnNetworksSuffix)
	vpn := env.VpnClient()

	t.Run("external requires cluster", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), vcnNetworkRequestTimeout)
		defer cancel()

		req := newExternalVcnNetworkRequest("single-external", hs.apiURL, hs.apiKey)
		_, err := vpn.CreateNetwork(ctx, req)
		require.Equal(t, codes.FailedPrecondition, status.Code(err), "%v", err)
		require.Equal(t, "external networks require cluster mode", status.Convert(err).Message())
	})

	enableStatefullPg(t, env)

	t.Run("external rejects bad token", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), vcnNetworkRequestTimeout)
		defer cancel()

		req := newExternalVcnNetworkRequest("cluster-external", hs.apiURL, "invalid.vcn-e2e-token")
		_, err := vpn.CreateNetwork(ctx, req)
		require.Equal(t, codes.Internal, status.Code(err), "%v", err)
		require.Contains(t, status.Convert(err).Message(), "listing namespaces, got status: 401")

		require.Empty(t, listVcnNetworks(ctx, t, vpn))
	})

	t.Run("lifecycle", func(t *testing.T) {
		testVcnNetworkLifecycle(t, env)
	})
}

func testVcnNetworkLifecycle(t *testing.T, env *TestEnvironment) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), vcnNetworkLifecycleTimeout)
	defer cancel()

	db := env.clusterDb(t)
	vpn := env.VpnClient()
	dockerClient := env.DockerClient()
	name := GetServiceName(t)
	created := deployVcnNetwork(ctx, t, vpn, name)
	getReq := &velez_api.GetVcnNetwork_Request{Id: created.GetId()}
	dropReq := &velez_api.DropVcnNetwork_Request{Id: created.GetId()}

	var isDropped bool
	t.Cleanup(func() {
		if isDropped {
			return
		}

		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), vcnNetworkRequestTimeout)
		defer cleanupCancel()

		_, err := vpn.DropNetwork(cleanupCtx, dropReq)
		if status.Code(err) != codes.NotFound {
			require.NoError(t, err)
		}
	})

	const storedContainerQuery = `SELECT container_id FROM velez.networks WHERE id = $1`

	var containerId string
	err := db.QueryRowContext(ctx, storedContainerQuery, created.GetId()).Scan(&containerId)
	require.NoError(t, err, "the deployed network must be stored in cluster Postgres")
	require.NotEmpty(t, containerId)

	inspected, err := dockerClient.ContainerInspect(ctx, containerId)
	require.NoError(t, err)
	require.NotNil(t, inspected.State)
	require.True(t, inspected.State.Running, "the deployed Headscale container must be running")

	listed := listVcnNetworks(ctx, t, vpn)
	require.Len(t, listed, 1)
	requireVcnNetworkDetails(t, created, listed[0])

	details, err := vpn.GetNetwork(ctx, getReq)
	require.NoError(t, err)
	requireVcnNetworkDetails(t, created, details.GetNetwork())
	require.Empty(t, details.GetPeers())

	verifyVcnNetworkKeyLifecycle(ctx, t, vpn, created, getReq)
	connectAndVerifyVcnNetworkPeer(ctx, t, env, created, getReq, name+"-app")

	_, err = vpn.DropNetwork(ctx, dropReq)
	require.NoError(t, err)
	isDropped = true

	_, err = dockerClient.ContainerInspect(ctx, containerId)
	require.True(t, errdefs.IsNotFound(err), "the Headscale container must be removed, got %v", err)

	err = db.QueryRowContext(ctx, storedContainerQuery, created.GetId()).Scan(&containerId)
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.Empty(t, listVcnNetworks(ctx, t, vpn))
}

func newExternalVcnNetworkRequest(name, url, token string) *velez_api.CreateVcnNetwork_Request {
	return &velez_api.CreateVcnNetwork_Request{
		Name: name,
		Source: &velez_api.CreateVcnNetwork_Request_External{
			External: &velez_api.CreateVcnNetwork_External{Url: url, Token: token},
		},
	}
}

func newDeployedVcnNetworkRequest(name string) *velez_api.CreateVcnNetwork_Request {
	return &velez_api.CreateVcnNetwork_Request{
		Name: name,
		Source: &velez_api.CreateVcnNetwork_Request_Deploy{
			Deploy: &velez_api.CreateVcnNetwork_Deploy{},
		},
	}
}

func deployVcnNetwork(
	ctx context.Context, t *testing.T, vpn velez_api.VcnApiClient, name string,
) *velez_api.VcnNetwork {
	t.Helper()

	gateway, err := dindGateway()
	require.NoError(t, err)

	// The default gRPC authority is the Velez container, while Headscale's
	// published port is reachable through the Docker host's gateway.
	deployCtx := metadata.AppendToOutgoingContext(ctx, "x-forwarded-host", gateway)
	req := newDeployedVcnNetworkRequest(name)
	resp, err := vpn.CreateNetwork(deployCtx, req)
	require.NoError(t, err)

	network := resp.GetNetwork()
	require.NotEmpty(t, network.GetId())
	require.Equal(t, name, network.GetName())
	require.Equal(t, "headscale", network.GetProvider())
	require.Equal(t, "deployed", network.GetSource())
	require.Equal(t, "vcn-"+name, network.GetServiceName())
	require.NotEmpty(t, network.GetUrl())
	require.NotNil(t, network.GetCreatedAt())

	return network
}

func listVcnNetworks(ctx context.Context, t *testing.T, vpn velez_api.VcnApiClient) []*velez_api.VcnNetwork {
	t.Helper()

	req := &velez_api.ListVcnNetworks_Request{}
	response, err := vpn.ListNetworks(ctx, req)
	require.NoError(t, err)

	return response.GetNetworks()
}

func requireVcnNetworkDetails(t *testing.T, expected, actual *velez_api.VcnNetwork) {
	t.Helper()

	require.Equal(t, expected.GetId(), actual.GetId())
	require.Equal(t, expected.GetName(), actual.GetName())
	require.Equal(t, expected.GetProvider(), actual.GetProvider())
	require.Equal(t, expected.GetSource(), actual.GetSource())
	require.Equal(t, expected.GetUrl(), actual.GetUrl())
	require.Equal(t, expected.GetServiceName(), actual.GetServiceName())
}

func verifyVcnNetworkKeyLifecycle(
	ctx context.Context, t *testing.T, vpn velez_api.VcnApiClient,
	network *velez_api.VcnNetwork, getReq *velez_api.GetVcnNetwork_Request,
) {
	t.Helper()

	issueReq := &velez_api.IssueNetworkApiKey_Request{NetworkId: network.GetId()}
	issued, err := vpn.IssueNetworkApiKey(ctx, issueReq)
	require.NoError(t, err)

	key := issued.GetKey()
	prefix := issued.GetApiKey().GetPrefix()
	require.NotEmpty(t, key)
	require.NotEmpty(t, prefix)
	require.True(t, strings.HasPrefix(key, prefix+"."), "the issue response must pair the token with its prefix")

	before, err := vpn.GetNetwork(ctx, getReq)
	require.NoError(t, err)

	apiKey := findVcnApiKey(before.GetApiKeys(), prefix)
	require.NotNil(t, apiKey, "GetNetwork must list metadata for the issued key")
	require.NotNil(t, apiKey.GetExpiration())
	require.True(t, apiKey.GetExpiration().AsTime().After(time.Now()))
	require.Equal(t, http.StatusOK, vcnHeadscaleAuthStatus(ctx, t, network.GetUrl(), key))

	expireReq := &velez_api.ExpireNetworkApiKey_Request{NetworkId: network.GetId(), Prefix: prefix}
	_, err = vpn.ExpireNetworkApiKey(ctx, expireReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, vcnHeadscaleAuthStatus(ctx, t, network.GetUrl(), key))

	after, err := vpn.GetNetwork(ctx, getReq)
	require.NoError(t, err)

	apiKey = findVcnApiKey(after.GetApiKeys(), prefix)
	if apiKey != nil {
		require.NotNil(t, apiKey.GetExpiration())
		require.True(t, apiKey.GetExpiration().AsTime().Before(time.Now()), "listed metadata must mark the key expired")
	}
}

func findVcnApiKey(keys []*velez_api.NetworkApiKey, prefix string) *velez_api.NetworkApiKey {
	for _, key := range keys {
		if key.GetPrefix() == prefix {
			return key
		}
	}

	return nil
}

func vcnHeadscaleAuthStatus(ctx context.Context, t *testing.T, loginUrl, key string) int {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, vcnNetworkRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, loginUrl+"/api/v1/user", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := dindHttpClient().Do(req)
	require.NoError(t, err)
	defer common.CloseWithLog(resp.Body.Close, "Headscale key check response")

	return resp.StatusCode
}

func connectAndVerifyVcnNetworkPeer(
	ctx context.Context, t *testing.T, env *TestEnvironment, network *velez_api.VcnNetwork,
	getReq *velez_api.GetVcnNetwork_Request, serviceName string,
) {
	t.Helper()

	smerd := newHelloWorldRequest(serviceName)
	env.CreateSmerd(t, smerd)

	vpn := env.VpnClient()
	dockerClient := env.DockerClient()
	sidecarName := serviceName + "-" + patterns.TailscaleSidecarSuffix
	peerHost := strings.ReplaceAll(sidecarName, "_", "-")
	networkId := network.GetId()

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), vcnNetworkRequestTimeout)
		defer cancel()

		removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}
		err := dockerClient.ContainerRemove(cleanupCtx, sidecarName, removeOpts)
		if !errdefs.IsNotFound(err) {
			require.NoError(t, err)
		}

		err = dockerClient.VolumeRemove(cleanupCtx, sidecarName, true)
		if !errdefs.IsNotFound(err) {
			require.NoError(t, err)
		}
	})

	req := newVcnConnectServiceRequest(serviceName, networkId)
	_, err := vpn.ConnectService(ctx, req)
	require.NoError(t, err)

	sidecar, err := dockerClient.ContainerInspect(ctx, sidecarName)
	require.NoError(t, err)
	require.NotNil(t, sidecar.State)
	require.True(t, sidecar.State.Running)
	require.NotNil(t, sidecar.Config)

	var loginArg string
	for _, envVar := range sidecar.Config.Env {
		value, found := strings.CutPrefix(envVar, "TS_EXTRA_ARGS=")
		if found {
			loginArg = value

			break
		}
	}

	require.Equal(t, "--login-server="+network.GetUrl(), loginArg)

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		pollCtx, cancel := context.WithTimeout(ctx, vcnNetworkRequestTimeout)
		defer cancel()

		details, err := vpn.GetNetwork(pollCtx, getReq)
		require.NoError(c, err)
		require.Len(c, details.GetPeers(), 1)

		peer := details.GetPeers()[0]
		require.Equal(c, peerHost, peer.GetHost())
		require.NotEmpty(c, peer.GetId())
		require.NotEmpty(c, peer.GetIpAddresses())
	}, vpnSidecarJoinTimeout, time.Second, "the sidecar must register in the selected named network")
}

func newVcnConnectServiceRequest(serviceName, networkId string) *velez_api.ConnectService_Request {
	return &velez_api.ConnectService_Request{
		ServiceName: serviceName,
		NetworkId:   &networkId,
	}
}
