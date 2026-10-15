//go:build e2e_full

package e2e

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	transferImage = "alpine"

	// An image whose container may join another's network namespace under the
	// runtime policy (the tailscale sidecar in production); alpine is tagged
	// under that name as a stand-in.
	transferSidecarImage = "tailscale/tailscale:e2e"

	transferSeedText   = "hello transfer"
	transferBlobBytes  = "3000000"
	transferDataDir    = "/data"
	transferReaderDir  = "/v"
	transferSeedScript = "head -c " + transferBlobBytes + " /dev/urandom > " + transferDataDir + "/blob.bin" +
		" && printf %s '" + transferSeedText + "' > " + transferDataDir + "/seed.txt && sleep 3600"
	transferDigestScript = "cat " + transferReaderDir + "/seed.txt; echo; sha256sum " + transferReaderDir + "/blob.bin"

	transferAwaitTimeout = 3 * time.Minute
	transferPollInterval = 500 * time.Millisecond
)

var transferSleepCommand = []string{"sleep", "3600"}

type transferPair struct {
	source *TestEnvironment
	target *TestEnvironment
}

func newTransferPair(t *testing.T, plane Plane) transferPair {
	t.Helper()

	targetHost := transferTargetDockerHost(t)

	source := plane.NewEnvironment(t)
	target := NewEnvironment(t, WithDockerHost(targetHost), WithTcpServer())

	return transferPair{source: source, target: target}
}

func transferClient(env *TestEnvironment) velez_api.TransferAPIClient {
	return velez_api.NewTransferAPIClient(env.grpcConn)
}

func newCreateTransferTokenRequest() *velez_api.CreateTransferToken_Request {
	return &velez_api.CreateTransferToken_Request{}
}

func newPlanTransferRequest(containerName string) *velez_api.PlanTransfer_Request {
	return &velez_api.PlanTransfer_Request{ContainerName: containerName}
}

func newStartTransferRequest(
	containerName string, sidecarNames []string, address, token string, isConsumersAcknowledged bool,
) *velez_api.StartTransfer_Request {
	return &velez_api.StartTransfer_Request{
		ContainerName:           containerName,
		SidecarNames:            sidecarNames,
		Address:                 address,
		Token:                   token,
		IsConsumersAcknowledged: isConsumersAcknowledged,
	}
}

func newListTransfersRequest() *velez_api.ListTransfers_Request {
	return &velez_api.ListTransfers_Request{}
}

func (p transferPair) issueToken(t *testing.T) string {
	t.Helper()

	resp, err := transferClient(p.target).CreateTransferToken(t.Context(), newCreateTransferTokenRequest())
	require.NoError(t, err)
	require.NotEmpty(t, resp.GetToken())

	return resp.GetToken()
}

func (p transferPair) plan(t *testing.T, containerName string) *velez_api.PlanTransfer_Response {
	t.Helper()

	resp, err := transferClient(p.source).PlanTransfer(t.Context(), newPlanTransferRequest(containerName))
	require.NoError(t, err)

	return resp
}

func (p transferPair) start(
	t *testing.T, containerName string, sidecarNames []string, token string, isConsumersAcknowledged bool,
) (*velez_api.StartTransfer_Response, error) {
	t.Helper()

	req := newStartTransferRequest(containerName, sidecarNames, p.target.TcpAddress(), token, isConsumersAcknowledged)

	return transferClient(p.source).StartTransfer(t.Context(), req)
}

func (p transferPair) awaitStatus(t *testing.T, transferId string, want velez_api.TransferStatus) {
	t.Helper()

	deadline := time.Now().Add(transferAwaitTimeout)

	for time.Now().Before(deadline) {
		info := p.findTransfer(t, transferId)

		if info.GetStatus() == want {
			return
		}

		isTerminal := info.GetStatus() == velez_api.TransferStatus_TRANSFER_STATUS_FAILED ||
			info.GetStatus() == velez_api.TransferStatus_TRANSFER_STATUS_CANCELLED
		require.False(t, isTerminal, "transfer %s settled as %s: %s; target saw: %s",
			transferId, info.GetStatus(), info.GetError(), p.targetErrors(t))

		time.Sleep(transferPollInterval)
	}

	require.Failf(t, "transfer did not settle", "transfer %s did not reach %s", transferId, want)
}

func (p transferPair) targetErrors(t *testing.T) []string {
	t.Helper()

	resp, err := transferClient(p.target).ListTransfers(t.Context(), newListTransfersRequest())
	require.NoError(t, err)

	errs := make([]string, 0, len(resp.GetTransfers()))

	for _, info := range resp.GetTransfers() {
		errs = append(errs, info.GetStatus().String()+": "+info.GetError())
	}

	return errs
}

func (p transferPair) findTransfer(t *testing.T, transferId string) *velez_api.TransferInfo {
	t.Helper()

	resp, err := transferClient(p.source).ListTransfers(t.Context(), newListTransfersRequest())
	require.NoError(t, err)

	for _, info := range resp.GetTransfers() {
		if info.GetId() == transferId {
			return info
		}
	}

	require.Failf(t, "transfer missing", "transfer %s is not listed", transferId)

	return nil
}

func (p transferPair) moveSucceeds(
	t *testing.T, containerName string, sidecarNames []string, token string, isConsumersAcknowledged bool,
) {
	t.Helper()

	resp, err := p.start(t, containerName, sidecarNames, token, isConsumersAcknowledged)
	require.NoError(t, err)

	p.awaitStatus(t, resp.GetTransferId(), velez_api.TransferStatus_TRANSFER_STATUS_DONE)
}

func dockerOf(env *TestEnvironment) client.APIClient {
	return env.Custom.NodeClients.Docker().Client()
}

func startFixtureContainer(
	t *testing.T,
	env *TestEnvironment,
	name string,
	cfg *container.Config,
	hostCfg *container.HostConfig,
) container.InspectResponse {
	t.Helper()

	cli := dockerOf(env)

	if cfg.Labels == nil {
		cfg.Labels = map[string]string{}
	}

	cfg.Labels[testCaseNameLabel] = t.Name()

	created, err := cli.ContainerCreate(t.Context(), cfg, hostCfg, nil, nil, name)
	require.NoError(t, err)

	err = cli.ContainerStart(t.Context(), created.ID, container.StartOptions{})
	require.NoError(t, err)

	inspected, err := cli.ContainerInspect(t.Context(), created.ID)
	require.NoError(t, err)

	return inspected
}

// createVolumeRoot starts a container that seeds a random blob and a text file
// into the named volume, on the named network ("" is the default bridge).
func createVolumeRoot(
	t *testing.T, env *TestEnvironment, name, volumeName, networkName string,
) container.InspectResponse {
	t.Helper()

	cfg := &container.Config{
		Image: transferImage,
		Cmd:   []string{"sh", "-c", transferSeedScript},
	}

	hostCfg := &container.HostConfig{
		Binds:       []string{volumeName + ":" + transferDataDir},
		NetworkMode: container.NetworkMode(networkName),
	}

	return startFixtureContainer(t, env, name, cfg, hostCfg)
}

func createBindRoot(t *testing.T, env *TestEnvironment, name, hostPath string) container.InspectResponse {
	t.Helper()

	cfg := &container.Config{
		Image: transferImage,
		Cmd:   transferSleepCommand,
	}

	hostCfg := &container.HostConfig{
		Binds: []string{hostPath + ":" + transferDataDir},
	}

	return startFixtureContainer(t, env, name, cfg, hostCfg)
}

func createNetworkConsumer(t *testing.T, env *TestEnvironment, name, networkName string) container.InspectResponse {
	t.Helper()

	cfg := &container.Config{
		Image: transferImage,
		Cmd:   transferSleepCommand,
	}

	hostCfg := &container.HostConfig{
		NetworkMode: container.NetworkMode(networkName),
	}

	return startFixtureContainer(t, env, name, cfg, hostCfg)
}

func createTransferSidecar(t *testing.T, env *TestEnvironment, name, rootId string) container.InspectResponse {
	t.Helper()

	err := dockerOf(env).ImageTag(t.Context(), transferImage, transferSidecarImage)
	require.NoError(t, err)

	cfg := &container.Config{
		Image:  transferSidecarImage,
		Cmd:    transferSleepCommand,
		Labels: map[string]string{labels.Sidecar: labelValueTrue},
	}

	hostCfg := &container.HostConfig{
		NetworkMode: container.NetworkMode(networkGroupNetworkPrefix + rootId),
	}

	return startFixtureContainer(t, env, name, cfg, hostCfg)
}

func createUserNetwork(t *testing.T, env *TestEnvironment, name string) {
	t.Helper()

	opts := network.CreateOptions{Labels: map[string]string{testCaseNameLabel: t.Name()}}

	_, err := dockerOf(env).NetworkCreate(t.Context(), name, opts)
	require.NoError(t, err)

	t.Cleanup(func() {
		removeTransferNetwork(env, name)
	})
}

func removeTransferNetwork(env *TestEnvironment, name string) {
	cli := dockerOf(env)

	list, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return
	}

	for _, c := range list {
		_, isAttached := c.NetworkSettings.Networks[name]
		if isAttached {
			_ = cli.ContainerRemove(context.Background(), c.ID, container.RemoveOptions{Force: true})
		}
	}

	_ = cli.NetworkRemove(context.Background(), name)
}

// volumeDigest reads the seeded text and the blob's checksum back out of a
// volume through a throwaway reader container on the env's daemon.
func volumeDigest(t *testing.T, env *TestEnvironment, volumeName string) string {
	t.Helper()

	cli := dockerOf(env)

	cfg := &container.Config{
		Image: transferImage,
		Cmd:   []string{"sh", "-c", transferDigestScript},
	}

	hostCfg := &container.HostConfig{
		Binds: []string{volumeName + ":" + transferReaderDir + ":ro"},
	}

	created, err := cli.ContainerCreate(t.Context(), cfg, hostCfg, nil, nil, "")
	require.NoError(t, err)

	defer func() {
		_ = cli.ContainerRemove(context.Background(), created.ID, container.RemoveOptions{Force: true})
	}()

	err = cli.ContainerStart(t.Context(), created.ID, container.StartOptions{})
	require.NoError(t, err)

	waitCh, errCh := cli.ContainerWait(t.Context(), created.ID, container.WaitConditionNotRunning)

	select {
	case <-waitCh:
	case waitErr := <-errCh:
		require.NoError(t, waitErr)
	}

	logs, err := cli.ContainerLogs(t.Context(), created.ID, container.LogsOptions{ShowStdout: true})
	require.NoError(t, err)

	defer func() {
		_ = logs.Close()
	}()

	var stdout, stderr bytes.Buffer

	_, err = stdcopy.StdCopy(&stdout, &stderr, logs)
	require.NoError(t, err)

	return stdout.String()
}

func awaitSeededVolume(t *testing.T, env *TestEnvironment, volumeName string) string {
	t.Helper()

	var digest string

	isSeeded := func() bool {
		digest = volumeDigest(t, env, volumeName)

		return strings.HasPrefix(digest, transferSeedText)
	}

	require.Eventually(t, isSeeded, transferAwaitTimeout, transferPollInterval, "volume %s was never seeded", volumeName)

	return digest
}

func inspectOnEnv(t *testing.T, env *TestEnvironment, nameOrId string) container.InspectResponse {
	t.Helper()

	inspected, err := dockerOf(env).ContainerInspect(t.Context(), nameOrId)
	require.NoError(t, err)

	return inspected
}

func requireStopped(t *testing.T, env *TestEnvironment, name string) container.InspectResponse {
	t.Helper()

	inspected := inspectOnEnv(t, env, name)
	require.False(t, inspected.State.Running, "container %s must not be running", name)

	return inspected
}

func requireRunning(t *testing.T, env *TestEnvironment, name string) {
	t.Helper()

	inspected := inspectOnEnv(t, env, name)
	require.True(t, inspected.State.Running, "container %s must be running", name)
}

func transferNames(t *testing.T) (rootName, volumeName string) {
	t.Helper()

	base := GetServiceName(t)

	return base + "_root", base + "_vol"
}

func createCollidingContainer(t *testing.T, env *TestEnvironment, name string) {
	t.Helper()

	ensureImagePresent(t, dockerOf(env), transferImage)

	cfg := &container.Config{
		Image:  transferImage,
		Cmd:    transferSleepCommand,
		Labels: map[string]string{testCaseNameLabel: t.Name()},
	}

	hostCfg := &container.HostConfig{}

	_, err := dockerOf(env).ContainerCreate(t.Context(), cfg, hostCfg, nil, nil, name)
	require.NoError(t, err)
}
