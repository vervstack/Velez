// Package dind spins up a disposable Docker-in-Docker daemon for the e2e
// suite: one privileged `docker:dind` container per test binary, created
// from a bootstrap socket (a local unix socket or a remote tcp:// host) and
// force-removed when the suite finishes.
//
// Every container the suite starts then lands inside that daemon, so
// parallel git worktrees on one machine never share Docker state: the DinD
// container is named after the checked-out branch (one per worktree, stable
// across runs) and its published ports are assigned by the bootstrap
// daemon, so nothing has to be coordinated between runs. Setup force-removes
// a leftover instance of the same name first, so a run that died without
// Teardown self-heals instead of leaking a container per crash.
package dind

import (
	"context"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
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

	// EnvName overrides the DinD container name key (by default the current
	// git branch). Set it when the branch is not a good per-worktree key -
	// a CI job id, say.
	EnvName = "VELEZ_E2E_DIND_NAME"

	// EnvPrepareOnly set to 1 brings the daemon up, seeds images and networks,
	// then exits without running tests; used to warm the cache volume.
	EnvPrepareOnly = "VELEZ_E2E_PREPARE"

	// EnvActive is exported by the harness (see e2e TestMain) once the
	// daemon is up, so helpers can refuse to run the suite against anything
	// but a disposable DinD.
	EnvActive = "VELEZ_E2E_DIND"

	defaultImage  = "docker:28-dind"
	daemonPort    = 2375
	labelHarness  = "velez.e2e.dind"
	namePrefix    = "velez-dind-"
	readyTimeout  = 45 * time.Second
	readyInterval = 500 * time.Millisecond

	removeRetryAttempts = 5
	removeRetryInterval = 500 * time.Millisecond

	pullAttempts     = 4
	pullRetryBackoff = 2 * time.Second

	// cacheVolumeSuffix names the persistent /var/lib/docker volume for a
	// given DinD container. It is deliberately NOT removed on Teardown so
	// warm runs reuse pulled image layers instead of re-fetching them. Stale
	// ones (unused for more than 3 days) are collected at the next Setup.
	cacheVolumeSuffix = "-cache"
)

var (
	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errDindDaemonPortNotPublished = rerrors.New("dind daemon port was not published")
	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errContainerNoNetworkSettings = rerrors.New("container has no network settings")

	transientPullMarkers = []string{
		"TLS handshake timeout",
		"i/o timeout",
		"connection reset",
		"EOF",
		"429 Too Many",
		"toomanyrequests",
		"500 Internal",
		"502 Bad Gateway",
		"503 Service",
		"504 Gateway",
		"status code 5",
	}
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

	// Name is the container name key. Empty falls back to EnvName, then the
	// current git branch, then the working directory's base name. The
	// resolved key is sanitized and prefixed with "velez-dind-".
	Name string

	// Publish lists container-side ports (besides the daemon port) to
	// publish on the bootstrap host, so the test process can reach services
	// started inside the DinD. Each maps to an ephemeral host port; read
	// the mapping back with Env.Addr.
	Publish []int

	// NameSuffix is appended to the resolved container name key, so a second
	// daemon of the same run (a transfer target) gets its own container and
	// cache volume.
	NameSuffix string
}

// Env is a running DinD daemon. Call Teardown when the suite finishes.
type Env struct {
	// DockerHost is the tcp:// address of the DinD daemon, ready to be
	// assigned to DOCKER_HOST.
	DockerHost string

	// BootstrapHost is the docker host the DinD container itself runs on, so a
	// second daemon can be started beside this one instead of inside it.
	BootstrapHost string

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
	name := resolveName(ctx, opts.Name) + opts.NameSuffix
	cacheVolume := name + cacheVolumeSuffix

	collectStaleCaches(ctx, bootstrap, cacheVolume)

	env, err := bringUp(ctx, bootstrap, img, name, dindHost, cacheVolume, opts.Publish)
	if err == nil {
		env.BootstrapHost = firstNonEmpty(bootstrapHost, client.DefaultDockerHost)

		return env, nil
	}

	// A half-written image-cache volume (a run killed mid-pull) can wedge the
	// daemon on start. Drop it once and retry from a clean cache.
	firstErr := err

	rmErr := bootstrap.VolumeRemove(context.Background(), cacheVolume, true)
	if rmErr != nil {
		return nil, rerrors.Wrap(firstErr, "dind bring-up failed; cache-volume cleanup also failed: "+rmErr.Error())
	}

	env, err = bringUp(ctx, bootstrap, img, name, dindHost, cacheVolume, opts.Publish)
	if err != nil {
		return nil, rerrors.Wrap(err, "error starting dind container after cache-volume reset")
	}

	env.BootstrapHost = firstNonEmpty(bootstrapHost, client.DefaultDockerHost)

	return env, nil
}

// bringUp removes any stale container of the same name, ensures the
// persistent image-cache volume exists, starts one privileged DinD instance
// bound to it, and waits for the daemon to answer.
func bringUp(
	ctx context.Context,
	bootstrap *client.Client,
	img, name, dindHost, cacheVolume string,
	publish []int,
) (*Env, error) {
	err := removeStale(ctx, bootstrap, name)
	if err != nil {
		return nil, rerrors.Wrap(err, "error removing stale dind container")
	}

	err = ensureCacheVolume(ctx, bootstrap, cacheVolume)
	if err != nil {
		return nil, rerrors.Wrap(err, "error ensuring dind image-cache volume")
	}

	containerId, err := runDind(ctx, bootstrap, img, name, dindHost, cacheVolume, withSocksPort(publish))
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

		return nil, errDindDaemonPortNotPublished
	}

	env.DockerHost = "tcp://" + net.JoinHostPort(dindHost, daemonHostPort)

	err = waitForDaemon(ctx, env.DockerHost)
	if err != nil {
		_ = env.Teardown()

		return nil, rerrors.Wrap(err, "dind daemon did not become ready")
	}

	err = env.startSocksProxy(ctx)
	if err != nil {
		_ = env.Teardown()

		return nil, rerrors.Wrap(err, "error starting socks proxy")
	}

	return env, nil
}

// ensureCacheVolume creates the named /var/lib/docker volume for a DinD
// container if it is absent. VolumeCreate returns the existing volume
// unchanged when one already exists, so this is idempotent across runs.
func ensureCacheVolume(ctx context.Context, cli *client.Client, name string) error {
	createOpts := volume.CreateOptions{
		Name: name,
		Labels: map[string]string{
			labelHarness: "true",
		},
	}

	_, err := cli.VolumeCreate(ctx, createOpts)
	if err != nil {
		return rerrors.Wrap(err, "error creating volume "+name)
	}

	return nil
}

// Reset removes every container, non-built-in network and volume from the
// DinD daemon, so each run starts from a clean inner state. The cache volume
// keeps /var/lib/docker across runs, which would otherwise carry the previous
// run's containers over. Dangling images are pruned; tagged and seeded ones
// stay.
func (e *Env) Reset(ctx context.Context) error {
	cli, err := newDockerClient(e.DockerHost)
	if err != nil {
		return rerrors.Wrap(err, "error creating dind docker client")
	}

	defer func() {
		_ = cli.Close()
	}()

	err = removeAllContainers(ctx, cli)
	if err != nil {
		return rerrors.Wrap(err, "error removing dind containers")
	}

	err = removeAllNetworks(ctx, cli)
	if err != nil {
		return rerrors.Wrap(err, "error removing dind networks")
	}

	err = removeAllVolumes(ctx, cli)
	if err != nil {
		return rerrors.Wrap(err, "error removing dind volumes")
	}

	pruneFilters := filters.NewArgs(filters.Arg("dangling", "true"))

	_, err = cli.ImagesPrune(ctx, pruneFilters)
	if err != nil {
		return rerrors.Wrap(err, "error pruning dangling images")
	}

	return nil
}

func removeAllContainers(ctx context.Context, cli *client.Client) error {
	listOpts := container.ListOptions{All: true}

	containers, err := cli.ContainerList(ctx, listOpts)
	if err != nil {
		return rerrors.Wrap(err, "error listing containers")
	}

	removeOpts := container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}

	for _, c := range containers {
		err = removeContainerRetrying(ctx, cli, c.ID, removeOpts)
		if err != nil {
			return rerrors.Wrap(err, "error removing container "+c.ID)
		}
	}

	return nil
}

func removeContainerRetrying(
	ctx context.Context,
	cli *client.Client,
	containerId string,
	removeOpts container.RemoveOptions,
) error {
	var err error

	for range removeRetryAttempts {
		err = cli.ContainerRemove(ctx, containerId, removeOpts)
		if err == nil || errdefs.IsNotFound(err) {
			return nil
		}

		if !strings.Contains(err.Error(), "is already in progress") {
			return rerrors.Wrap(err, "error removing container")
		}

		select {
		case <-ctx.Done():
			return rerrors.Wrap(ctx.Err(), "error waiting for container removal")
		case <-time.After(removeRetryInterval):
		}
	}

	return rerrors.Wrap(err, "error removing container, removal still in progress")
}

func removeAllNetworks(ctx context.Context, cli *client.Client) error {
	listOpts := network.ListOptions{}

	networks, err := cli.NetworkList(ctx, listOpts)
	if err != nil {
		return rerrors.Wrap(err, "error listing networks")
	}

	for _, n := range networks {
		if isBuiltInNetwork(n.Name) {
			continue
		}

		err = cli.NetworkRemove(ctx, n.ID)
		if err != nil && !errdefs.IsNotFound(err) {
			return rerrors.Wrap(err, "error removing network "+n.Name)
		}
	}

	return nil
}

func isBuiltInNetwork(name string) bool {
	return name == "bridge" || name == "host" || name == "none"
}

func removeAllVolumes(ctx context.Context, cli *client.Client) error {
	listOpts := volume.ListOptions{}

	volumes, err := cli.VolumeList(ctx, listOpts)
	if err != nil {
		return rerrors.Wrap(err, "error listing volumes")
	}

	for _, v := range volumes.Volumes {
		err = cli.VolumeRemove(ctx, v.Name, true)
		if err != nil && !errdefs.IsNotFound(err) {
			return rerrors.Wrap(err, "error removing volume "+v.Name)
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

// Teardown force-removes the DinD container and closes the bootstrap
// client. The named image-cache volume ("<name>-cache") is deliberately
// left in place so the next run reuses pulled image layers; RemoveVolumes
// only drops anonymous volumes. Safe to call more than once.
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
	isPresent, err := hasImage(ctx, cli, ref)
	if err != nil {
		return rerrors.Wrap(err, "error checking image")
	}

	if isPresent {
		return nil
	}

	backoff := pullRetryBackoff

	for attempt := 1; ; attempt++ {
		err = pullImage(ctx, cli, ref)
		if err == nil {
			return nil
		}

		if attempt == pullAttempts || !isTransientPullError(err) {
			return rerrors.Wrap(err, "error pulling image")
		}

		select {
		case <-ctx.Done():
			return rerrors.Wrap(ctx.Err(), "error waiting to retry pull")
		case <-time.After(backoff):
		}

		backoff *= 2
	}
}

func pullImage(ctx context.Context, cli *client.Client, ref string) error {
	pullOpts := image.PullOptions{}

	reader, err := cli.ImagePull(ctx, ref, pullOpts)
	if err != nil {
		return rerrors.Wrap(err, "error starting pull")
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

func isTransientPullError(err error) bool {
	if errdefs.IsUnavailable(err) {
		return true
	}

	message := err.Error()

	for _, marker := range transientPullMarkers {
		if strings.Contains(message, marker) {
			return true
		}
	}

	return false
}

func runDind(
	ctx context.Context,
	cli *client.Client,
	img, name, dindHost, cacheVolume string,
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

	cacheMount := mount.Mount{
		Type:   mount.TypeVolume,
		Source: cacheVolume,
		Target: "/var/lib/docker",
	}

	hostCfg := &container.HostConfig{
		Privileged:   true,
		PortBindings: bindings,
		Mounts:       []mount.Mount{cacheMount},
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
		return nil, errContainerNoNetworkSettings
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

// resolveName builds the DinD container name: a sanitized key (the explicit
// option, else EnvName, else the git branch, else the working directory
// name) behind the "velez-dind-" prefix. Stable for one worktree across
// runs so Setup can force-remove the previous instance rather than pile up
// one container per run.
func resolveName(ctx context.Context, explicit string) string {
	key := sanitizeName(firstNonEmpty(explicit, os.Getenv(EnvName), gitBranch(ctx), currentDirBase()))

	if key == "" {
		key = strconv.Itoa(os.Getpid())
	}

	return namePrefix + key
}

func gitBranch(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")

	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	branch := strings.TrimSpace(string(out))
	if branch == "HEAD" {
		return ""
	}

	return branch
}

func currentDirBase() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	return filepath.Base(wd)
}

// sanitizeName lower-cases raw and maps every run of characters Docker
// disallows in a container name to a single '-', trimming separators from
// the ends.
func sanitizeName(raw string) string {
	var b strings.Builder

	for _, r := range strings.ToLower(raw) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}

	return strings.Trim(b.String(), "-_.")
}

// removeStale force-removes a leftover DinD container of the same name so a
// run that never reached Teardown does not block the next one. The named
// image-cache volume is left intact. A missing container is not an error.
func removeStale(ctx context.Context, cli *client.Client, name string) error {
	removeOpts := container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}

	err := cli.ContainerRemove(ctx, name, removeOpts)
	if err != nil && !errdefs.IsNotFound(err) {
		return rerrors.Wrap(err, "error removing container "+name)
	}

	return nil
}
