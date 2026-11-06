package selfswap

import (
	"context"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	// apiContainerPort is the container port Velez serves its API on; the
	// health check hits it on the host port it is bound to.
	apiContainerPort = "53890/tcp"

	// stopTimeout bounds how long the swapper waits for the target (and,
	// during rollback, the broken replacement) to stop before forcing.
	stopTimeout = 30
	// httpClientTimeout bounds a single health-check request.
	httpClientTimeout = time.Second * 2

	// rollbackOpTimeout detaches rollback operations from the caller's
	// context: a cancelled swap must still restore a running Velez, and
	// every rollback call would otherwise fail immediately with
	// context.Canceled. It exceeds stopTimeout so a replacement that
	// ignores SIGTERM can still be SIGKILLed and the old container
	// restored within this window.
	rollbackOpTimeout = time.Second * 90
)

type settings struct {
	targetContainerId string
	// image is the image the new container is created from;
	// empty means the target's own image.
	image          string
	healthTimeout  time.Duration
	rollbackWindow time.Duration
}

func settingsFromEnv() (settings, error) {
	targetId := os.Getenv(EnvTargetContainerId)
	if targetId == "" {
		return settings{}, rerrors.New("target container id env is empty: " + EnvTargetContainerId)
	}

	st := settings{
		targetContainerId: targetId,
		image:             os.Getenv(EnvImage),
		healthTimeout:     DefaultHealthTimeout,
		rollbackWindow:    DefaultRollbackWindow,
	}

	healthTimeoutSec := os.Getenv(EnvHealthTimeout)
	if healthTimeoutSec != "" {
		healthTimeout, err := strconv.Atoi(healthTimeoutSec)
		if err != nil {
			return settings{}, rerrors.Wrap(err, "error parsing "+EnvHealthTimeout)
		}

		st.healthTimeout = time.Duration(healthTimeout) * time.Second
	}

	rollbackWindowSec := os.Getenv(EnvRollbackWindow)
	if rollbackWindowSec != "" {
		rollbackWindow, err := strconv.Atoi(rollbackWindowSec)
		if err != nil {
			return settings{}, rerrors.Wrap(err, "error parsing "+EnvRollbackWindow)
		}

		st.rollbackWindow = time.Duration(rollbackWindow) * time.Second
	}

	return st, nil
}

type swapper struct {
	cli      client.APIClient
	log      zerolog.Logger
	settings settings

	// checkHealth is a field so docker-backed tests can stub the /api/version
	// polling against throwaway dummy containers.
	checkHealth func(ctx context.Context, hostPort string, timeout time.Duration) error
}

// Run executes the self-swap sequence in a minimal container: no servers, no
// storage - it stops the target Velez container, renames it to the rollback
// name, recreates Velez from the requested image, health-checks it and rolls
// back on failure.
//
// It is invoked from main() when EnvSelfSwap=1, before the regular app is
// built; the swapper container only mounts docker.sock.
func Run(ctx context.Context) error {
	st, err := settingsFromEnv()
	if err != nil {
		return rerrors.Wrap(err, "error reading swapper settings")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return rerrors.Wrap(err, "error building docker client")
	}

	defer func() {
		_ = cli.Close()
	}()

	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().Timestamp().Str("component", SwapperContainerName).Logger()

	sw := &swapper{
		cli:         cli,
		log:         log,
		settings:    st,
		checkHealth: checkApiVersion,
	}

	return sw.run(ctx)
}

func (s *swapper) run(ctx context.Context) error {
	inspect, err := s.cli.ContainerInspect(ctx, s.settings.targetContainerId)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting target container "+s.settings.targetContainerId)
	}

	originalName := containerName(inspect.Name)
	s.log.Info().Str("target", originalName).Msg("target container found")

	err = s.removeStaleRollback(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error removing stale rollback container")
	}

	imageRef := s.settings.image
	if imageRef == "" {
		imageRef = inspect.Config.Image
	}

	err = s.ensureImage(ctx, imageRef)
	if err != nil {
		return rerrors.Wrap(err, "error ensuring image "+imageRef)
	}

	// Clone the config before stopping: values are read from the inspect
	// result, which was fetched above.
	newConfig, newHostConfig, newNetConfig := cloneConfig(&inspect, imageRef)

	s.log.Info().Msg("stopping target container")

	err = s.stopContainer(ctx, inspect.ID)
	if err != nil {
		return rerrors.Wrap(err, "error stopping target container")
	}

	err = s.cli.ContainerRename(ctx, inspect.ID, RollbackContainerName)
	if err != nil {
		return rerrors.Wrap(err, "error renaming target to rollback container")
	}

	s.log.Info().Str("image", imageRef).Msg("creating new container")

	created, err := s.cli.ContainerCreate(ctx, newConfig, newHostConfig, newNetConfig, nil, originalName)
	if err != nil {
		s.rollback(&inspect, originalName)

		return rerrors.Wrap(err, "error creating new container")
	}

	err = s.cli.ContainerStart(ctx, created.ID, container.StartOptions{})
	if err != nil {
		s.rollback(&inspect, originalName)

		return rerrors.Wrap(err, "error starting new container")
	}

	hostPort := hostApiPort(newHostConfig)

	s.log.Info().Str("port", hostPort).Msg("waiting for new container to become healthy")

	err = s.checkHealth(ctx, hostPort, s.settings.healthTimeout)
	if err != nil {
		s.log.Error().Err(err).Msg("health check failed, rolling back")
		s.rollback(&inspect, originalName)

		return rerrors.Wrap(err, "health check failed")
	}

	s.log.Info().Msg("swap successful")

	err = s.waitRollbackWindow(ctx)
	if err != nil {
		// The new container is already running and healthy - the node is
		// not at risk; the stale rollback container is cleaned up by the
		// next swap.
		return rerrors.Wrap(err, "error waiting out rollback window")
	}

	err = s.dropRollback(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error dropping rollback container")
	}

	// Self-removal is safe because the removal request travels over the
	// mounted docker.sock and outlives this process. It runs only on the
	// success path: on failure the swapper must stay as an exited
	// container so its logs remain readable.
	err = s.removeSelf(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error removing swapper container")
	}

	return nil
}

// rollback restores the old container under its original name after the new
// one failed to come up. It runs detached from the swap context so a
// cancelled swap still leaves a running Velez behind.
func (s *swapper) rollback(inspect *container.InspectResponse, originalName string) {
	ctx, cancel := context.WithTimeout(context.Background(), rollbackOpTimeout)
	defer cancel()

	s.log.Info().Msg("rolling back to previous container")

	newContainer, inspectErr := s.cli.ContainerInspect(ctx, originalName)
	if inspectErr == nil {
		err := s.stopContainer(ctx, newContainer.ID)
		if err != nil {
			s.log.Error().Err(err).Msg("error stopping broken replacement")
		}

		err = s.cli.ContainerRemove(ctx, newContainer.ID, container.RemoveOptions{Force: true})
		if err != nil {
			s.log.Error().Err(err).Msg("error removing broken replacement")
		}
	}

	err := s.cli.ContainerRename(ctx, inspect.ID, originalName)
	if err != nil {
		s.log.Error().Err(err).Msg("error renaming rollback container back")

		return
	}

	err = s.cli.ContainerStart(ctx, inspect.ID, container.StartOptions{})
	if err != nil {
		s.log.Error().Err(err).Msg("error restarting previous container")
	}
}

func (s *swapper) stopContainer(ctx context.Context, containerId string) error {
	noRestart := container.UpdateConfig{
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyDisabled,
		},
	}

	_, err := s.cli.ContainerUpdate(ctx, containerId, noRestart)
	if err != nil {
		return rerrors.Wrap(err, "error disabling restart policy")
	}

	stopTimeoutSec := stopTimeout
	stopOptions := container.StopOptions{
		Timeout: &stopTimeoutSec,
	}

	err = s.cli.ContainerStop(ctx, containerId, stopOptions)
	if err != nil {
		return rerrors.Wrap(err, "error stopping container")
	}

	return nil
}

func (s *swapper) removeStaleRollback(ctx context.Context) error {
	_, err := s.cli.ContainerInspect(ctx, RollbackContainerName)
	if err == nil {
		s.log.Info().Msg("removing stale rollback container")

		removeOptions := container.RemoveOptions{
			Force: true,
		}

		err = s.cli.ContainerRemove(ctx, RollbackContainerName, removeOptions)
		if err != nil {
			return rerrors.Wrap(err, "error removing rollback container")
		}

		return nil
	}

	// No stale rollback container - nothing to do. The daemon returns
	// a "No such container" error for both missing names and ids.
	return nil
}

func (s *swapper) ensureImage(ctx context.Context, imageRef string) error {
	_, err := s.cli.ImageInspect(ctx, imageRef)
	if err == nil {
		return nil
	}

	s.log.Info().Str("image", imageRef).Msg("image not present locally, pulling")

	pullOptions := image.PullOptions{}

	pullResp, err := s.cli.ImagePull(ctx, imageRef, pullOptions)
	if err != nil {
		return rerrors.Wrap(err, "error pulling image")
	}

	defer func() {
		_ = pullResp.Close()
	}()

	_, err = io.Copy(io.Discard, pullResp)
	if err != nil {
		return rerrors.Wrap(err, "error reading image pull response")
	}

	return nil
}

// waitRollbackWindow keeps the rollback container around so a failed new
// container can still be rolled back manually, then lets run drop it.
func (s *swapper) waitRollbackWindow(ctx context.Context) error {
	s.log.Info().Dur("window", s.settings.rollbackWindow).Msg("waiting out rollback window")

	timer := time.NewTimer(s.settings.rollbackWindow)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return rerrors.Wrap(ctx.Err(), "swap context cancelled during rollback window")
	case <-timer.C:
		return nil
	}
}

func (s *swapper) dropRollback(ctx context.Context) error {
	s.log.Info().Msg("dropping rollback container")

	removeOptions := container.RemoveOptions{
		Force: true,
	}

	err := s.cli.ContainerRemove(ctx, RollbackContainerName, removeOptions)
	if err != nil {
		return rerrors.Wrap(err, "error removing rollback container")
	}

	return nil
}

func (s *swapper) removeSelf(ctx context.Context) error {
	selfId := env.GetContainerId()
	if selfId == nil {
		s.log.Warn().Msg("own container id is unknown, skipping self-removal")

		return nil
	}

	s.log.Info().Str("id", *selfId).Msg("removing swapper container")

	removeOptions := container.RemoveOptions{
		Force: true,
	}

	err := s.cli.ContainerRemove(ctx, *selfId, removeOptions)
	if err != nil {
		return rerrors.Wrap(err, "error removing own container")
	}

	return nil
}

// cloneConfig builds the create-request configs for the replacement
// container from the target's inspect result, swapping in the requested
// image and stripping swapper-specific leftovers.
func cloneConfig(inspect *container.InspectResponse, imageRef string) (
	*container.Config, *container.HostConfig, *network.NetworkingConfig,
) {
	newConfig := *inspect.Config

	newConfig.Image = imageRef

	newConfig.Env = make([]string, 0, len(inspect.Config.Env))
	for _, e := range inspect.Config.Env {
		if strings.HasPrefix(e, EnvSelfSwap) {
			continue
		}

		newConfig.Env = append(newConfig.Env, e)
	}

	newConfig.Labels = make(map[string]string, len(inspect.Config.Labels))
	for k, v := range inspect.Config.Labels {
		// Docker container labels are immutable, so the swapper cannot
		// mark the rollback container with RollbackLabel either - both
		// labels are stripped defensively from the clone instead.
		if k == labels.SwapperLabel || k == labels.RollbackLabel {
			continue
		}

		newConfig.Labels[k] = v
	}

	newHostConfig := *inspect.HostConfig
	newNetConfig := cloneNetworking(inspect)

	return &newConfig, &newHostConfig, newNetConfig
}

// cloneNetworking rebuilds the endpoint configs from the inspect result.
// MacAddress and IPAMConfig are deliberately dropped: the old container is
// still attached to the networks (only stopped) at create time, so cloning
// either would collide with it.
func cloneNetworking(inspect *container.InspectResponse) *network.NetworkingConfig {
	if inspect.NetworkSettings == nil {
		return nil
	}

	endpoints := make(map[string]*network.EndpointSettings, len(inspect.NetworkSettings.Networks))
	for name, nw := range inspect.NetworkSettings.Networks {
		if nw == nil {
			continue
		}

		endpoints[name] = &network.EndpointSettings{
			Links:   nw.Links,
			Aliases: nw.Aliases,
		}
	}

	if len(endpoints) == 0 {
		return nil
	}

	return &network.NetworkingConfig{
		EndpointsConfig: endpoints,
	}
}

func containerName(inspectName string) string {
	return strings.TrimPrefix(inspectName, "/")
}

// hostApiPort resolves the host port the new container binds for the API
// container port, falling back to the bare container port.
func hostApiPort(hostConfig *container.HostConfig) string {
	bindings, ok := hostConfig.PortBindings[nat.Port(apiContainerPort)]
	if !ok || len(bindings) == 0 || bindings[0].HostPort == "" {
		return "53890"
	}

	return bindings[0].HostPort
}

func checkApiVersion(ctx context.Context, hostPort string, timeout time.Duration) error {
	url := "http://127.0.0.1:" + hostPort + "/api/version"
	httpClient := &http.Client{
		Timeout: httpClientTimeout,
	}

	deadline := time.Now().Add(timeout)

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return rerrors.Wrap(err, "error building health check request")
		}

		resp, err := httpClient.Do(req)
		if err == nil {
			closeErr := resp.Body.Close()
			if closeErr != nil {
				return rerrors.Wrap(closeErr, "error closing health check response")
			}

			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		if time.Now().After(deadline) {
			return rerrors.New("health check timed out after " + timeout.String())
		}

		select {
		case <-ctx.Done():
			return rerrors.Wrap(ctx.Err(), "health check aborted")
		case <-time.After(time.Second):
		}
	}
}
