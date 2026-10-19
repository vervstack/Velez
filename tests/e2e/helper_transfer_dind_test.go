//go:build e2e_full

package e2e

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	dockernet "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	clusterenv "go.vervstack.ru/Velez/internal/cluster/env"
)

const (
	transferDindName  = "e2e-transfer-dind2"
	transferDindImage = "docker:28-dind"
	transferDindPort  = "2375"

	transferDindReadyTimeout = 90 * time.Second
	transferDindReadyPoll    = time.Second
	transferDindPingTimeout  = 3 * time.Second
)

var (
	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errTransferDindNotReady = rerrors.New("transfer dind did not become ready")

	transferDaemonMu     sync.Mutex     //nolint:gochecknoglobals // one target daemon per test process
	transferDaemonCached *managedDaemon //nolint:gochecknoglobals // outcome of the one-time start
)

// WithManagedDaemon points the environment's Velez at daemon instead of the
// DinD it runs in.
func WithManagedDaemon(daemon *managedDaemon) TestEnvOpt {
	return func(a *TestEnvironment) {
		a.managed = daemon
	}
}

// transferDaemonFor is the transfer target's daemon: one privileged DinD
// started inside the shared DinD, prepared once per test process.
func transferDaemonFor(t *testing.T) (*managedDaemon, error) {
	t.Helper()

	transferDaemonMu.Lock()
	defer transferDaemonMu.Unlock()

	if transferDaemonCached != nil {
		return transferDaemonCached, nil
	}

	daemon, err := startTransferDaemon(context.Background())
	if err != nil {
		return nil, rerrors.Wrap(err, "error starting the transfer target daemon")
	}

	transferDaemonCached = daemon

	return daemon, nil
}

func startTransferDaemon(ctx context.Context) (*managedDaemon, error) {
	outer, err := client.NewClientWithOpts(client.WithHost(sharedDind.DockerHost), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating the shared dind client")
	}

	defer func() { _ = outer.Close() }()

	ip, err := runTransferDindContainer(ctx, outer)
	if err != nil {
		return nil, rerrors.Wrap(err, "error running the target dind container")
	}

	inner, err := client.NewClientWithOpts(
		client.WithHost("tcp://"+net.JoinHostPort(ip, transferDindPort)),
		client.WithDialContext(dindDialIp),
		client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating the target dind client")
	}

	err = awaitTransferDind(ctx, inner)
	if err != nil {
		return nil, rerrors.Wrap(err, "error waiting for the target dind")
	}

	err = prepareTransferDind(ctx, outer, inner)
	if err != nil {
		return nil, rerrors.Wrap(err, "error preparing the target dind")
	}

	daemon := &managedDaemon{name: transferDindName, port: transferDindPort, client: inner}

	return daemon, nil
}

func runTransferDindContainer(ctx context.Context, outer *client.Client) (string, error) {
	removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}

	err := outer.ContainerRemove(ctx, transferDindName, removeOpts)
	if err != nil && !errdefs.IsNotFound(err) {
		return "", rerrors.Wrap(err, "error removing a stale target dind")
	}

	cfg := &container.Config{
		Image: transferDindImage,
		Env:   []string{"DOCKER_TLS_CERTDIR="},
	}

	hostCfg := &container.HostConfig{
		Privileged:  true,
		NetworkMode: container.NetworkMode(clusterenv.VervNetwork),
	}

	netCfg := &dockernet.NetworkingConfig{
		EndpointsConfig: map[string]*dockernet.EndpointSettings{clusterenv.VervNetwork: {}},
	}

	created, err := outer.ContainerCreate(ctx, cfg, hostCfg, netCfg, nil, transferDindName)
	if err != nil {
		return "", rerrors.Wrap(err, "error creating the target dind container")
	}

	err = outer.ContainerStart(ctx, created.ID, container.StartOptions{})
	if err != nil {
		return "", rerrors.Wrap(err, "error starting the target dind container")
	}

	inspected, err := outer.ContainerInspect(ctx, created.ID)
	if err != nil {
		return "", rerrors.Wrap(err, "error inspecting the target dind container")
	}

	endpoint, ok := inspected.NetworkSettings.Networks[clusterenv.VervNetwork]
	if !ok || endpoint.IPAddress == "" {
		return "", rerrors.Wrap(errDindNoContainerIp, transferDindName)
	}

	return endpoint.IPAddress, nil
}

func awaitTransferDind(ctx context.Context, inner *client.Client) error {
	ctx, cancel := context.WithTimeout(ctx, transferDindReadyTimeout)
	defer cancel()

	var lastErr error

	for {
		pingCtx, pingCancel := context.WithTimeout(ctx, transferDindPingTimeout)

		_, lastErr = inner.Ping(pingCtx)

		pingCancel()

		if lastErr == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return rerrors.Wrap(errTransferDindNotReady, lastErr.Error())
		case <-time.After(transferDindReadyPoll):
		}
	}
}

// prepareTransferDind gives the target daemon what the moved containers need:
// the base image (also tagged as the tailscale sidecar stand-in) and the
// networks Velez expects on a node.
func prepareTransferDind(ctx context.Context, outer, inner *client.Client) error {
	err := copyTransferImage(ctx, outer, inner, transferImage)
	if err != nil {
		return rerrors.Wrap(err, "error copying the image into the target dind")
	}

	err = inner.ImageTag(ctx, transferImage, transferSidecarImage)
	if err != nil {
		return rerrors.Wrap(err, "error tagging the sidecar image")
	}

	for _, name := range dindEnsureNetworks {
		createOpts := dockernet.CreateOptions{}

		_, err = inner.NetworkCreate(ctx, name, createOpts)
		if err != nil {
			return rerrors.Wrap(err, "error creating network "+name)
		}
	}

	return nil
}

func copyTransferImage(ctx context.Context, from, to *client.Client, ref string) error {
	archive, err := from.ImageSave(ctx, []string{ref})
	if err != nil {
		return rerrors.Wrap(err, "error saving image")
	}

	defer func() { _ = archive.Close() }()

	loadResp, err := to.ImageLoad(ctx, archive, client.ImageLoadWithQuiet(true))
	if err != nil {
		return rerrors.Wrap(err, "error loading image")
	}

	defer func() { _ = loadResp.Body.Close() }()

	_, err = io.Copy(io.Discard, loadResp.Body)
	if err != nil {
		return rerrors.Wrap(err, "error draining image load log")
	}

	return nil
}
