// Package dind spins up a disposable Docker-in-Docker daemon for the e2e
// suite: one privileged `docker:dind` container per test binary, created
// from a bootstrap socket (a local unix socket or a remote tcp:// host) and
// force-removed when the suite finishes.
//
// Every container the suite starts then lands inside that daemon, so
// parallel git worktrees on one machine never share Docker state: the DinD
// container name is unique per process and its published ports are assigned
// by the bootstrap daemon, so nothing has to be coordinated between runs.
package dind

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.redsock.ru/rerrors"
)

const (
	// EnvBootstrapHost overrides the docker host the DinD container itself
	// is created on. Unset means client.FromEnv (the ambient DOCKER_HOST or
	// the local unix socket).
	EnvBootstrapHost = "VELEZ_E2E_DOCKER_HOST"

	// EnvImage overrides the docker:dind image reference.
	EnvImage = "VELEZ_E2E_DIND_IMAGE"

	// EnvActive is exported by the harness (see e2e TestMain) once the
	// daemon is up, so helpers can refuse to run the suite against anything
	// but a disposable DinD.
	EnvActive = "VELEZ_E2E_DIND"

	defaultImage  = "docker:28-dind"
	daemonPort    = 2375
	labelHarness  = "velez.e2e.dind"
	readyTimeout  = 45 * time.Second
	readyInterval = 500 * time.Millisecond
)

// Options is the knob set for Setup. The zero value is valid.
type Options struct {
	// BootstrapHost is the docker host the DinD container is created on.
	// Empty falls back to EnvBootstrapHost, then the ambient DOCKER_HOST,
	// then the local unix socket.
	BootstrapHost string

	// Image is the docker:dind image reference. Empty falls back to
	// EnvImage, then defaultImage.
	Image string

	// Publish lists container-side ports (besides the daemon port) to
	// publish on the bootstrap host, so the test process can reach services
	// started inside the DinD. Each maps to an ephemeral host port; read
	// the mapping back with Env.Addr.
	Publish []int
}

// Env is a running DinD daemon. Call Teardown when the suite finishes.
type Env struct {
	// DockerHost is the tcp:// address of the DinD daemon, ready to be
	// assigned to DOCKER_HOST.
	DockerHost string

	bootstrap   *client.Client
	containerId string
	dindHost    string
	ports       map[int]string // container-side port -> host-side port
}

// Setup pulls the dind image if needed, starts one privileged instance on
// the bootstrap host, waits for its daemon to answer, and returns a handle.
func Setup(ctx context.Context, opts Options) (*Env, error) {
	bootstrapHost := firstNonEmpty(opts.BootstrapHost, os.Getenv(EnvBootstrapHost), os.Getenv("DOCKER_HOST"))

	bootstrap, err := newDockerClient(bootstrapHost)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating bootstrap docker client")
	}

	img := firstNonEmpty(opts.Image, os.Getenv(EnvImage), defaultImage)

	err = ensureImage(ctx, bootstrap, img)
	if err != nil {
		return nil, rerrors.Wrap(err, "error pulling dind image")
	}

	dindHost := resolveDindHost(bootstrapHost)
	name := fmt.Sprintf("velez-dind-%d-%d", os.Getpid(), time.Now().UnixNano())

	containerId, err := runDind(ctx, bootstrap, img, name, dindHost, opts.Publish)
	if err != nil {
		return nil, rerrors.Wrap(err, "error starting dind container")
	}

	env := &Env{
		bootstrap:   bootstrap,
		containerId: containerId,
		dindHost:    dindHost,
	}

	env.ports, err = readPublishedPorts(ctx, bootstrap, containerId)
	if err != nil {
		teardownErr := env.Teardown()
		if teardownErr != nil {
			err = rerrors.Wrap(err, "teardown also failed: "+teardownErr.Error())
		}

		return nil, rerrors.Wrap(err, "error reading published ports")
	}

	daemonHostPort, ok := env.ports[daemonPort]
	if !ok {
		_ = env.Teardown()

		return nil, rerrors.New("dind daemon port was not published")
	}

	env.DockerHost = "tcp://" + net.JoinHostPort(dindHost, daemonHostPort)

	err = waitForDaemon(ctx, env.DockerHost)
	if err != nil {
		_ = env.Teardown()

		return nil, rerrors.Wrap(err, "dind daemon did not become ready")
	}

	return env, nil
}

// Seed pulls the given image references into the DinD daemon unless they
// are already present. The e2e suite calls this for images that some code
// paths create a container from without pulling first (the cluster postgres
// pattern, for one) - that only ever worked against a warm host daemon.
func (e *Env) Seed(ctx context.Context, refs ...string) error {
	cli, err := newDockerClient(e.DockerHost)
	if err != nil {
		return rerrors.Wrap(err, "error creating dind docker client")
	}

	defer func() {
		_ = cli.Close()
	}()

	for _, ref := range refs {
		err = ensureImage(ctx, cli, ref)
		if err != nil {
			return rerrors.Wrap(err, "error seeding image "+ref)
		}
	}

	return nil
}

// EnsureNetwork creates the named bridge networks in the DinD daemon if
// they are absent. Velez attaches port-exposing containers to a fixed
// "verv" network it no longer creates itself (env.StartNetwork is disabled)
// - a real node already has it, a fresh DinD does not.
func (e *Env) EnsureNetwork(ctx context.Context, names ...string) error {
	cli, err := newDockerClient(e.DockerHost)
	if err != nil {
		return rerrors.Wrap(err, "error creating dind docker client")
	}

	defer func() {
		_ = cli.Close()
	}()

	createOpts := network.CreateOptions{
		Driver: "bridge",
	}

	for _, name := range names {
		_, err = cli.NetworkCreate(ctx, name, createOpts)
		if err != nil && !errdefs.IsConflict(err) {
			return rerrors.Wrap(err, "error creating network "+name)
		}
	}

	return nil
}

// Addr returns the bootstrap-host address (host:port) a published
// container-side port is reachable at, and whether it was published.
func (e *Env) Addr(containerPort int) (string, bool) {
	hostPort, ok := e.ports[containerPort]
	if !ok {
		return "", false
	}

	return net.JoinHostPort(e.dindHost, hostPort), true
}

// Teardown force-removes the DinD container (and its anonymous
// /var/lib/docker volume) and closes the bootstrap client. Safe to call
// more than once.
func (e *Env) Teardown() error {
	if e.bootstrap == nil {
		return nil
	}

	removeOpts := container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}

	err := e.bootstrap.ContainerRemove(context.Background(), e.containerId, removeOpts)
	if err != nil && !errdefs.IsNotFound(err) {
		return rerrors.Wrap(err, "error removing dind container")
	}

	closeErr := e.bootstrap.Close()

	e.bootstrap = nil

	if closeErr != nil {
		return rerrors.Wrap(closeErr, "error closing bootstrap docker client")
	}

	return nil
}

func newDockerClient(host string) (*client.Client, error) {
	opts := []client.Opt{client.WithAPIVersionNegotiation()}

	if host == "" {
		opts = append(opts, client.FromEnv)
	} else {
		opts = append(opts, client.WithHost(host))
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating docker client")
	}

	return cli, nil
}

// resolveDindHost turns the bootstrap docker host into the host the test
// process should dial the DinD (and its published services) on: loopback
// for a local unix socket or a loopback tcp endpoint, the remote host
// otherwise.
func resolveDindHost(bootstrapHost string) string {
	const loopback = "127.0.0.1"

	if bootstrapHost == "" {
		return loopback
	}

	parsed, err := url.Parse(bootstrapHost)
	if err != nil {
		return loopback
	}

	switch parsed.Scheme {
	case "unix", "npipe", "":
		return loopback
	}

	host := parsed.Hostname()

	switch host {
	case "", "localhost", loopback, "0.0.0.0", "::1":
		return loopback
	}

	return host
}

func ensureImage(ctx context.Context, cli *client.Client, ref string) error {
	filterArgs := filters.NewArgs(filters.Arg("reference", ref))

	listOpts := image.ListOptions{
		Filters: filterArgs,
	}

	existing, err := cli.ImageList(ctx, listOpts)
	if err != nil {
		return rerrors.Wrap(err, "error listing images")
	}

	if len(existing) > 0 {
		return nil
	}

	pullOpts := image.PullOptions{}

	reader, err := cli.ImagePull(ctx, ref, pullOpts)
	if err != nil {
		return rerrors.Wrap(err, "error pulling image")
	}

	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		_ = reader.Close()

		return rerrors.Wrap(err, "error draining image pull log")
	}

	err = reader.Close()
	if err != nil {
		return rerrors.Wrap(err, "error closing image pull reader")
	}

	return nil
}

func runDind(
	ctx context.Context,
	cli *client.Client,
	img, name, dindHost string,
	publish []int,
) (string, error) {
	hostIp := ""
	if dindHost == "127.0.0.1" {
		hostIp = "127.0.0.1"
	}

	exposed := nat.PortSet{}
	bindings := nat.PortMap{}

	ports := make([]int, 0, 1+len(publish))

	ports = append(ports, daemonPort)
	ports = append(ports, publish...)

	for _, p := range ports {
		port, err := nat.NewPort("tcp", strconv.Itoa(p))
		if err != nil {
			return "", rerrors.Wrap(err, "error building port spec")
		}

		exposed[port] = struct{}{}

		binding := nat.PortBinding{
			HostIP:   hostIp,
			HostPort: "",
		}

		bindings[port] = []nat.PortBinding{binding}
	}

	// DOCKER_TLS_CERTDIR="" makes dind serve plain HTTP on 2375 instead of
	// generating certs and serving TLS on 2376.
	cfg := &container.Config{
		Image:        img,
		Env:          []string{"DOCKER_TLS_CERTDIR="},
		ExposedPorts: exposed,
		Labels: map[string]string{
			labelHarness: "true",
		},
	}

	hostCfg := &container.HostConfig{
		Privileged:   true,
		PortBindings: bindings,
	}

	created, err := cli.ContainerCreate(ctx, cfg, hostCfg, nil, nil, name)
	if err != nil {
		return "", rerrors.Wrap(err, "error creating container")
	}

	startOpts := container.StartOptions{}

	err = cli.ContainerStart(ctx, created.ID, startOpts)
	if err != nil {
		return "", rerrors.Wrap(err, "error starting container")
	}

	return created.ID, nil
}

func readPublishedPorts(ctx context.Context, cli *client.Client, containerId string) (map[int]string, error) {
	inspected, err := cli.ContainerInspect(ctx, containerId)
	if err != nil {
		return nil, rerrors.Wrap(err, "error inspecting container")
	}

	if inspected.NetworkSettings == nil {
		return nil, rerrors.New("container has no network settings")
	}

	result := make(map[int]string, len(inspected.NetworkSettings.Ports))

	for port, portBindings := range inspected.NetworkSettings.Ports {
		if len(portBindings) == 0 {
			continue
		}

		result[port.Int()] = portBindings[0].HostPort
	}

	return result, nil
}

func waitForDaemon(ctx context.Context, dockerHost string) error {
	cli, err := newDockerClient(dockerHost)
	if err != nil {
		return rerrors.Wrap(err, "error creating dind docker client")
	}

	defer func() {
		_ = cli.Close()
	}()

	deadline := time.Now().Add(readyTimeout)

	var lastErr error

	for time.Now().Before(deadline) {
		_, lastErr = cli.Ping(ctx)
		if lastErr == nil {
			return nil
		}

		time.Sleep(readyInterval)
	}

	return rerrors.Wrap(lastErr, "timed out waiting for dind daemon")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}
