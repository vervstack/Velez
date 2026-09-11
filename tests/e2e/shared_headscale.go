package e2e

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// sharedHeadscale is the single, process-wide real headscale container used
// by every WithStateVcnEnabled() TestEnvironment in this package. It is
// owned entirely by the test harness: the Velez app under test connects to
// it via headscale.Connect(url, key) (the "existing headscale" branch of
// verv_closed_network.SetupVcn), so nothing here needs a product change to
// the app's own LaunchHeadscale path. Torn down by TestMain (main_test.go)
// after every test finishes.
//
// Scope of what this fixture proves: real namespace CRUD and pre-auth-key
// issuance through the VcnApi RPCs against a real headscale + real sqlite.
// It deliberately does NOT verify a tailnet join - the tailscale sidecar
// dials getLoginServerURLJob's hardcoded https://vcn.redsock.ru regardless
// of which headscale is actually running, so a hermetic join is not
// testable until that URL becomes configurable (Trello #126 follow-up).
var (
	sharedHeadscale         *sharedHeadscaleInstance
	initSharedHeadscaleOnce sync.Once
	errInitSharedHeadscale  error

	// headscaleAPIKeyPattern matches a headscale API key (prefix.secret)
	// inside the noisy multiplexed `docker exec` output.
	headscaleAPIKeyPattern = regexp.MustCompile(`[A-Za-z0-9]{6,}\.[A-Za-z0-9]{20,}`)

	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errHeadscalePortNotPublished = rerrors.New("dind did not publish the headscale port")
)

const (
	sharedHeadscaleContainerName = "e2e-shared-headscale"
	sharedHeadscaleImage         = "headscale/headscale:0.27.2-rc.1"
	sharedHeadscaleContainerPort = "8080/tcp"

	headscaleConfigMode      = 0o644
	headscaleConfigDirMode   = 0o755
	headscaleAPIKeyTimeout   = 60 * time.Second
	headscaleAPIReadyTimeout = 30 * time.Second
	headscalePollInterval    = time.Second

	// headscaleFixtureConfig is a minimal, self-contained headscale server
	// config: sqlite in an anonymous volume, the embedded DERP server
	// enabled (headscale refuses to start with an empty DERPMap), and no
	// external dependencies. server_url is a placeholder - the API surface
	// this fixture exercises (users/preauthkey) does not depend on it.
	headscaleFixtureConfig = `---
server_url: http://headscale.e2e:8080
listen_addr: 0.0.0.0:8080
metrics_listen_addr: 127.0.0.1:9090
noise:
  private_key_path: /var/lib/headscale/noise_private.key
prefixes:
  v4: 100.64.0.0/10
  v6: fd7a:115c:a1e0::/48
  allocation: sequential
derp:
  server:
    enabled: true
    region_id: 999
    region_code: local
    region_name: local
    stun_listen_addr: "0.0.0.0:3478"
    private_key_path: /var/lib/headscale/derp_private.key
  urls: []
  paths: []
  auto_update_enabled: false
database:
  type: sqlite
  sqlite:
    path: /var/lib/headscale/db.sqlite
log:
  level: info
  format: text
dns:
  magic_dns: false
  override_local_dns: false
  nameservers:
    global: []
unix_socket: /var/run/headscale/headscale.sock
`
)

// sharedHeadscaleInstance is a handle to the running fixture container.
type sharedHeadscaleInstance struct {
	docker *docker.Docker

	containerID string

	// apiURL is the bootstrap-host address the in-process Velez app dials
	// (sharedDind.Addr-translated), NOT a DinD-internal address.
	apiURL string
	apiKey string
}

// getSharedHeadscale lazily creates the one headscale container shared by
// every WithStateVcnEnabled() TestEnvironment in this package, the first
// time any VPN test asks for it. Mirrors getSharedMatreshka's sync.Once
// pattern so -run filters that never touch VPN don't pay container spin-up
// cost.
func getSharedHeadscale(t *testing.T) *sharedHeadscaleInstance {
	t.Helper()

	initSharedHeadscaleOnce.Do(func() {
		sharedHeadscale, errInitSharedHeadscale = startSharedHeadscale(context.Background())
	})

	require.NoError(t, errInitSharedHeadscale, "shared headscale fixture must come up")

	return sharedHeadscale
}

func startSharedHeadscale(ctx context.Context) (*sharedHeadscaleInstance, error) {
	dockerClient, err := docker.NewClient(nil)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating docker client for headscale fixture")
	}

	api := dockerClient.Client()

	removeOpts := container.RemoveOptions{Force: true}

	// Drop a stale container from a crashed previous run before recreating.
	_ = api.ContainerRemove(ctx, sharedHeadscaleContainerName, removeOpts)

	hostPort, ok := sharedDind.Addr(dindHeadscalePort)
	if !ok {
		return nil, errHeadscalePortNotPublished
	}

	cfg := &container.Config{
		Image:        sharedHeadscaleImage,
		Hostname:     "headscale",
		Cmd:          []string{"serve"},
		ExposedPorts: nat.PortSet{sharedHeadscaleContainerPort: struct{}{}},
	}

	hostCfg := &container.HostConfig{
		PortBindings: nat.PortMap{
			sharedHeadscaleContainerPort: []nat.PortBinding{{HostPort: strconv.Itoa(dindHeadscalePort)}},
		},
	}

	netCfg := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			env.VervNetwork: {},
		},
	}

	created, err := api.ContainerCreate(ctx, cfg, hostCfg, netCfg, nil, sharedHeadscaleContainerName)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating headscale container")
	}

	err = api.CopyToContainer(ctx, created.ID, "/etc", headscaleConfigTar(), container.CopyToContainerOptions{})
	if err != nil {
		return nil, rerrors.Wrap(err, "error copying headscale config into container")
	}

	err = api.ContainerStart(ctx, created.ID, container.StartOptions{})
	if err != nil {
		return nil, rerrors.Wrap(err, "error starting headscale container")
	}

	instance := &sharedHeadscaleInstance{
		docker:      dockerClient,
		containerID: created.ID,
		apiURL:      "http://" + hostPort,
	}

	instance.apiKey, err = instance.awaitAPIKey(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error waiting for headscale api key")
	}

	err = instance.awaitAPIReady(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error waiting for headscale api")
	}

	return instance, nil
}

// awaitAPIKey polls `headscale apikey create` until headscale has finished
// booting and hands back a key. Mirrors the app's own
// headscale.issueNewAPIKey exec.
func (i *sharedHeadscaleInstance) awaitAPIKey(ctx context.Context) (string, error) {
	execCfg := container.ExecOptions{
		Cmd:          []string{"headscale", "apikey", "create"},
		Env:          []string{"HEADSCALE_LOG_FORMAT=text", "NO_COLOR=1"},
		AttachStdout: true,
	}

	deadline := time.Now().Add(headscaleAPIKeyTimeout)

	var lastErr error

	for time.Now().Before(deadline) {
		out, err := i.docker.Exec(ctx, i.containerID, execCfg)
		if err != nil {
			lastErr = rerrors.Wrap(err, "error exec-ing headscale apikey create")

			time.Sleep(headscalePollInterval)

			continue
		}

		key := headscaleAPIKeyPattern.FindString(string(out))
		if key != "" {
			return key, nil
		}

		lastErr = user_errors.New("headscale apikey output had no key: " + strings.TrimSpace(string(out)))

		time.Sleep(headscalePollInterval)
	}

	return "", lastErr
}

// awaitAPIReady confirms the HTTP API is reachable from this process at the
// bootstrap-host address the Velez app will use.
func (i *sharedHeadscaleInstance) awaitAPIReady(ctx context.Context) error {
	deadline := time.Now().Add(headscaleAPIReadyTimeout)

	var lastErr error

	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, i.apiURL+"/api/v1/user", nil)
		if err != nil {
			return rerrors.Wrap(err, "error building headscale readiness request")
		}

		req.Header.Set("Authorization", "Bearer "+i.apiKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = rerrors.Wrap(err, "error calling headscale api")

			time.Sleep(headscalePollInterval)

			continue
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		lastErr = user_errors.New("headscale /api/v1/user returned " + resp.Status)

		time.Sleep(headscalePollInterval)
	}

	return lastErr
}

func (i *sharedHeadscaleInstance) stop() {
	if i == nil {
		return
	}

	removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}

	err := i.docker.Client().ContainerRemove(context.Background(), i.containerID, removeOpts)
	if err != nil {
		log.Error().Err(err).Msg("error removing shared headscale container")
	}
}

// headscaleConfigTar builds a tar rooted at /etc (the CopyToContainer dst):
// a headscale/ dir entry plus headscale/config.yaml, so the mount point the
// image lacks is created along with the file.
func headscaleConfigTar() io.Reader {
	buf := &bytes.Buffer{}
	tw := tar.NewWriter(buf)

	dirHdr := &tar.Header{
		Name:     "headscale/",
		Mode:     headscaleConfigDirMode,
		Typeflag: tar.TypeDir,
	}

	_ = tw.WriteHeader(dirHdr)

	fileHdr := &tar.Header{
		Name: "headscale/config.yaml",
		Mode: headscaleConfigMode,
		Size: int64(len(headscaleFixtureConfig)),
	}

	_ = tw.WriteHeader(fileHdr)
	_, _ = tw.Write([]byte(headscaleFixtureConfig))
	_ = tw.Close()

	return bytes.NewReader(buf.Bytes())
}
