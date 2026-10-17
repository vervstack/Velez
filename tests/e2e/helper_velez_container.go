package e2e

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/local_state"
	clusterenv "go.vervstack.ru/Velez/internal/cluster/env"
)

// Container backend: NewEnvironment starts Velez as a container in the group's
// DinD instead of building it in-process, selected by VELEZ_E2E_BACKEND.
//
// A fixed pool of velezSlotCount slots is leased per test. Slot k owns the
// container name velez-e2e-<k> and the port slice
// [dindPortBandStart+k*velezSlotPortCount, +velezSlotPortCount-1]. With the
// group's SOCKS5 proxy the whole slice is the container's AvailablePorts and
// the test dials its gRPC port by container IP; without the proxy the first
// port publishes the gRPC port and the rest are AvailablePorts.
const (
	envBackend       = "VELEZ_E2E_BACKEND"
	backendContainer = "container"

	velezSlotCount     = 10
	velezSlotPortCount = 19

	velezContainerPort = nat.Port("53890/tcp")
	velezNamePrefix    = "velez-e2e-"
	velezSlotLabel     = testLabelPrefix + "velez_slot"

	velezDockerSocket = "/var/run/docker.sock"
	velezStateDir     = "/state"
	velezStateFile    = "state.json"
	velezOverrideFile = "/app/e2e_override.yaml"

	velezReadyTimeout = 90 * time.Second
	velezReadyPoll    = 300 * time.Millisecond
	velezReadyRpc     = 3 * time.Second
	velezLogTail      = "80"
	velezRestartGrace = 20

	velezDirMode  = 0o755
	velezFileMode = 0o644
)

var (
	velezSlots         chan int
	initVelezSlotsOnce sync.Once

	errVelezNotReady     = rerrors.New("velez container did not become ready")
	errVelezImageMissing = rerrors.New(envVelezImage + " is not set; run the suite through tests/e2egroups")
)

type velezContainer struct {
	slot    int
	name    string
	id      string
	gateway string
}

func isContainerBackend() bool {
	return os.Getenv(envBackend) == backendContainer
}

func velezSlotPool() chan int {
	initVelezSlotsOnce.Do(func() {
		velezSlots = make(chan int, velezSlotCount)

		for slot := range velezSlotCount {
			velezSlots <- slot
		}
	})

	return velezSlots
}

func acquireVelezSlot(t *testing.T) int {
	t.Helper()

	select {
	case slot := <-velezSlotPool():
		return slot
	case <-t.Context().Done():
		t.Fatal("no velez slot became free before the test was cancelled")

		return 0
	}
}

func velezSlotFirstPort(slot int) int {
	return dindPortBandStart + slot*velezSlotPortCount
}

func velezGrpcIsProxied() bool {
	_, ok := sharedDind.SocksAddr()

	return ok
}

func velezSlotAvailablePorts(slot int) []int {
	first := velezSlotFirstPort(slot)
	available := make([]int, 0, velezSlotPortCount)

	if velezGrpcIsProxied() {
		available = append(available, first)
	}

	for port := first + 1; port < first+velezSlotPortCount; port++ {
		available = append(available, port)
	}

	return available
}

// startContainerBackend replaces Custom.Init/Start and the bufconn server:
// it runs Velez from the e2e image and points env's gRPC connection at it.
func startContainerBackend(t *testing.T, env *TestEnvironment) {
	t.Helper()

	if env.dockerHost != "" || env.isTcpServer || env.configPath != defaultConfigPath {
		t.Skip("in-process only: a custom docker host, tcp server or config path")
	}

	image := os.Getenv(envVelezImage)
	require.NotEmpty(t, image, errVelezImageMissing.Error())

	slot := acquireVelezSlot(t)

	t.Cleanup(func() { velezSlotPool() <- slot })

	env.Cfg.Environment.AvailablePorts = velezSlotAvailablePorts(slot)

	vc := &velezContainer{slot: slot, name: velezNamePrefix + strconv.Itoa(slot)}

	env.velez = vc
	env.clean()

	removeVelezContainer(t, env, vc.name)

	vc.gateway = vervNetworkGateway(t, env)

	vc.id = createVelezContainer(t, env, image)

	t.Cleanup(func() {
		logVelezTailOnFailure(t, env)
		removeVelezContainer(t, env, vc.name)
	})

	t.Cleanup(env.clean)

	err := env.DockerClient().ContainerStart(t.Context(), vc.id, container.StartOptions{})
	require.NoError(t, err)

	connectVelezGrpc(t, env)
}

func createVelezContainer(t *testing.T, env *TestEnvironment, image string) string {
	t.Helper()

	vc := env.velez
	dockerClient := env.DockerClient()

	cfg := &container.Config{
		Image:        image,
		Env:          velezEnvVars(env),
		ExposedPorts: nat.PortSet{velezContainerPort: struct{}{}},
		Labels:       map[string]string{velezSlotLabel: strconv.Itoa(vc.slot)},
	}

	overrides := velezOverridesYaml(env)
	if overrides != "" {
		cfg.Cmd = []string{"-config", velezOverrideFile}
	}

	hostCfg := &container.HostConfig{
		Binds:       []string{velezDockerSocket + ":" + velezDockerSocket},
		NetworkMode: container.NetworkMode(clusterenv.VervNetwork),
	}

	if !velezGrpcIsProxied() {
		binding := nat.PortBinding{HostPort: strconv.Itoa(velezSlotFirstPort(vc.slot))}

		hostCfg.PortBindings = nat.PortMap{velezContainerPort: []nat.PortBinding{binding}}
	}

	netCfg := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{clusterenv.VervNetwork: {}},
	}

	created, err := dockerClient.ContainerCreate(t.Context(), cfg, hostCfg, netCfg, nil, vc.name)
	require.NoError(t, err)

	archive := velezInjectedFiles(t, env, overrides)

	err = dockerClient.CopyToContainer(t.Context(), created.ID, "/", archive, container.CopyToContainerOptions{})
	require.NoError(t, err)

	return created.ID
}

// velezEnvVars maps the loaded config onto the ENVIRONMENT_<KEY> variables
// the image's matreshka config reads. Shut-down-on-exit stays off: stopping the
// container is a SIGTERM that would drop the managed containers a restart must
// keep, and teardown removes them by the test label anyway.
func velezEnvVars(env *TestEnvironment) []string {
	cfg := env.Cfg.Environment

	labelList := strings.Join(cfg.CustomLabels, ",")

	vars := []string{
		"ENVIRONMENT_LOCAL_STATE_PATH=" + velezStateDir + "/" + velezStateFile,
		"ENVIRONMENT_AVAILABLE_PORTS=[" + portRange(cfg.AvailablePorts) + "]",
		"ENVIRONMENT_SHUT_DOWN_ON_EXIT=false",
		"ENVIRONMENT_CPU_DEFAULT=" + strconv.FormatFloat(cfg.CPUDefault, 'f', -1, 64),
		"ENVIRONMENT_MEMORY_SWAP_MB=" + strconv.Itoa(cfg.MemorySwapMb),
		"ENVIRONMENT_RAM_MB_DEFAULT=" + strconv.Itoa(cfg.RAMMbDefault),
		"ENVIRONMENT_AUTO_UPDATE_IS_ENABLED=" + strconv.FormatBool(cfg.AutoUpdateIsEnabled),
		"ENVIRONMENT_AUTO_UPDATE_CHECK_PERIOD_SECONDS=" + strconv.Itoa(cfg.AutoUpdateCheckPeriodSeconds),
		"ENVIRONMENT_MATRESHKA_IS_ENABLED=" + strconv.FormatBool(cfg.MatreshkaIsEnabled),
		"ENVIRONMENT_MATRESHKA_PORT=" + strconv.Itoa(cfg.MatreshkaPort),
		"ENVIRONMENT_CUSTOM_LABELS=[" + labelList + "]",
	}

	optional := map[string]string{
		"ENVIRONMENT_MATRESHKA_IMAGE":  cfg.MatreshkaImage,
		"ENVIRONMENT_CONTAINER_SUFFIX": cfg.ContainerSuffix,
		"ENVIRONMENT_CLUSTER_PG_DSN":   dindInternalAddrs(cfg.ClusterPgDsn, env.velez.gateway),
	}

	if cfg.MigrationsDir != "" {
		optional["ENVIRONMENT_MIGRATIONS_DIR"] = "/app/migrations"
	}

	if len(cfg.Environments) > 0 {
		optional["ENVIRONMENT_ENVIRONMENTS"] = "[" + strings.Join(cfg.Environments, ",") + "]"
	}

	for key, value := range optional {
		if value != "" {
			vars = append(vars, key+"="+value)
		}
	}

	return vars
}

func portRange(ports []int) string {
	if len(ports) == 0 {
		return ""
	}

	return strconv.Itoa(ports[0]) + ":" + strconv.Itoa(ports[len(ports)-1])
}

// velezOverridesYaml is the yaml-only service_discovery section: overrides
// registered by the options, with DinD-published host addresses rewritten to
// ones the Velez container can dial.
func velezOverridesYaml(env *TestEnvironment) string {
	overrides := env.Cfg.Overrides.Overrides
	if len(overrides) == 0 {
		return ""
	}

	var out strings.Builder

	out.WriteString("service_discovery:\n  overrides:\n")

	for _, override := range overrides {
		out.WriteString("    - service_name: " + override.ServiceName + "\n      urls:\n")

		for _, url := range override.Urls {
			out.WriteString("        - " + dindInternalAddrs(url, env.velez.gateway) + "\n")
		}
	}

	return out.String()
}

// velezInjectedFiles is a tar of the files that must exist before Velez
// starts: its state file and, when there are overrides, the -config yaml.
func velezInjectedFiles(t *testing.T, env *TestEnvironment, overridesYaml string) io.Reader {
	t.Helper()

	state := velezSeedState(t, env)

	var buf bytes.Buffer

	writer := tar.NewWriter(&buf)

	dirHeader := &tar.Header{Name: strings.TrimPrefix(velezStateDir, "/") + "/", Typeflag: tar.TypeDir, Mode: velezDirMode}

	err := writer.WriteHeader(dirHeader)
	require.NoError(t, err)

	addTarFile(t, writer, strings.TrimPrefix(velezStateDir, "/")+"/"+velezStateFile, state)

	if overridesYaml != "" {
		addTarFile(t, writer, strings.TrimPrefix(velezOverrideFile, "/"), []byte(overridesYaml))
	}

	err = writer.Close()
	require.NoError(t, err)

	return &buf
}

func addTarFile(t *testing.T, writer *tar.Writer, name string, content []byte) {
	t.Helper()

	header := &tar.Header{Name: name, Typeflag: tar.TypeReg, Mode: velezFileMode, Size: int64(len(content))}

	err := writer.WriteHeader(header)
	require.NoError(t, err)

	_, err = writer.Write(content)
	require.NoError(t, err)
}

// velezSeedState is the host-side state file the options wrote, with
// addresses of DinD-published fixtures rewritten to ones reachable from
// inside the DinD.
func velezSeedState(t *testing.T, env *TestEnvironment) []byte {
	t.Helper()

	raw, err := os.ReadFile(env.Cfg.Environment.LocalStatePath)
	require.NoError(t, err)

	var state local_state.State

	err = json.Unmarshal(raw, &state)
	require.NoError(t, err)

	state.Network.Headscale.ServerUrl = dindInternalAddrs(state.Network.Headscale.ServerUrl, env.velez.gateway)

	seeded, err := json.Marshal(state)
	require.NoError(t, err)

	return seeded
}

// dindInternalAddrs rewrites every bootstrap-host address the DinD publishes
// (sharedDind.Addr) to gateway:<container-side port>, the address a container
// inside the DinD reaches the same listener at.
func dindInternalAddrs(text, gateway string) string {
	if text == "" {
		return text
	}

	for _, port := range dindPublishPorts() {
		hostAddr, ok := sharedDind.Addr(port)
		if !ok {
			continue
		}

		text = strings.ReplaceAll(text, hostAddr, net.JoinHostPort(gateway, strconv.Itoa(port)))
	}

	return text
}

func vervNetworkGateway(t *testing.T, env *TestEnvironment) string {
	t.Helper()

	inspected, err := env.DockerClient().NetworkInspect(t.Context(), clusterenv.VervNetwork, network.InspectOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, inspected.IPAM.Config)

	return inspected.IPAM.Config[0].Gateway
}

func removeVelezContainer(t *testing.T, env *TestEnvironment, name string) {
	t.Helper()

	removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}

	err := env.DockerClient().ContainerRemove(context.Background(), name, removeOpts)
	if err != nil && !errdefs.IsNotFound(err) {
		require.NoError(t, err)
	}
}

func logVelezTailOnFailure(t *testing.T, env *TestEnvironment) {
	t.Helper()

	if !t.Failed() {
		return
	}

	logsOpts := container.LogsOptions{ShowStdout: true, ShowStderr: true, Tail: velezLogTail}

	logs, err := env.DockerClient().ContainerLogs(context.Background(), env.velez.id, logsOpts)
	if err != nil {
		t.Logf("velez container logs unavailable: %v", err)

		return
	}

	defer func() { _ = logs.Close() }()

	var out bytes.Buffer

	_, err = stdcopy.StdCopy(&out, &out, logs)
	if err != nil {
		t.Logf("velez container logs unreadable: %v", err)

		return
	}

	t.Logf("velez container %s log tail:\n%s", env.velez.name, out.String())
}

func (e *TestEnvironment) dindGrpcDialer() func(ctx context.Context, addr string) (net.Conn, error) {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		return e.DindDialContext(ctx, "tcp", addr)
	}
}

// connectVelezGrpc dials the container's gRPC port, through the DinD's SOCKS5
// proxy when it has one and through the published slot port otherwise, and
// waits until an authenticated unary call succeeds.
func connectVelezGrpc(t *testing.T, env *TestEnvironment) {
	t.Helper()

	if velezGrpcIsProxied() {
		target := net.JoinHostPort(env.velez.name, velezContainerPort.Port())

		connectGrpc(t, env, "passthrough:///"+target, grpc.WithContextDialer(env.dindGrpcDialer()))
	} else {
		hostAddr, ok := sharedDind.Addr(velezSlotFirstPort(env.velez.slot))
		require.True(t, ok, "dind did not publish the velez slot port")

		connectGrpc(t, env, "passthrough:///"+hostAddr)
	}

	ctx, cancel := context.WithTimeout(t.Context(), velezReadyTimeout)
	defer cancel()

	var lastErr error

	for {
		callCtx, callCancel := context.WithTimeout(ctx, velezReadyRpc)

		_, lastErr = env.VelezApiClient().Version(callCtx, &velez_api.Version_Request{})

		callCancel()

		if lastErr == nil {
			return
		}

		select {
		case <-ctx.Done():
			require.NoError(t, rerrors.Wrap(errVelezNotReady, lastErr.Error()))

			return
		case <-time.After(velezReadyPoll):
		}
	}
}
