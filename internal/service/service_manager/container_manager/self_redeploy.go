package container_manager

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/selfswap"
)

// dockerSocketContainerPath is where the swapper container expects the node's
// docker socket to be mounted (internal/selfswap only mounts docker.sock).
const (
	dockerSocketContainerPath = "/var/run/docker.sock"
)

// SelfRedeploy pulls the image and starts the velez-swapper container that
// carries out the actual stop/rename/recreate swap against the target Velez
// container (internal/selfswap). The jobs engine is drained after the pull
// and before the hand-off: a pull failure leaves the node fully operational,
// and once the swapper starts the old process may die at any moment.
func (c *ContainerServiceImpl) SelfRedeploy(ctx context.Context, image, targetContainerId string) error {
	if image == "" {
		inspect, err := c.dockerAPI.ContainerInspect(ctx, targetContainerId)
		if err != nil {
			return rerrors.Wrap(err, "error inspecting own container")
		}

		image = inspect.Config.Image
	}

	// force: a redeploy must re-pull so "same image" still gets the freshest tag.
	_, err := dockerutils.PullImage(ctx, c.dockerAPI, image, true)
	if err != nil {
		return rerrors.Wrap(err, "error pulling image for self redeploy")
	}

	err = c.removeStaleSwapper(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error removing stale swapper container")
	}

	err = c.drainJobs(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error draining jobs engine")
	}

	err = c.startSwapper(ctx, image, targetContainerId)
	if err != nil {
		return rerrors.Wrap(err, "error starting swapper container")
	}

	return nil
}

// drainJobs rejects new tasks and waits for the active ones to finish. The
// wait is best-effort: single-node in-memory tasks are lost on restart by
// design, so a timeout only logs and proceeds.
func (c *ContainerServiceImpl) drainJobs(ctx context.Context) error {
	if c.jobsEngine == nil {
		return nil
	}

	c.jobsEngine.BeginDrain()

	if c.jobsDrainTimeout <= 0 {
		return nil
	}

	drainCtx, cancel := context.WithTimeout(ctx, c.jobsDrainTimeout)
	defer cancel()

	err := c.jobsEngine.WaitForDrain(drainCtx)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("self redeploy: jobs engine drain did not finish in time, proceeding")
	}

	return nil
}

// removeStaleSwapper drops a swapper container left behind by an interrupted
// attempt; a missing one is fine.
func (c *ContainerServiceImpl) removeStaleSwapper(ctx context.Context) error {
	removeOptions := container.RemoveOptions{Force: true}

	err := c.dockerAPI.ContainerRemove(ctx, selfswap.SwapperContainerName, removeOptions)
	if err != nil && !errdefs.IsNotFound(err) {
		return rerrors.Wrap(err, "error removing swapper container")
	}

	return nil
}

func (c *ContainerServiceImpl) startSwapper(ctx context.Context, image, targetContainerId string) error {
	swapperConfig := &container.Config{
		Image: image,
		Env: []string{
			selfswap.EnvSelfSwap + "=1",
			selfswap.EnvTargetContainerId + "=" + targetContainerId,
			selfswap.EnvImage + "=" + image,
			selfswap.EnvHealthTimeout + "=" + strconv.FormatInt(int64(selfswap.DefaultHealthTimeout/time.Second), 10),
			selfswap.EnvRollbackWindow + "=" + strconv.FormatInt(int64(selfswap.DefaultRollbackWindow/time.Second), 10),
		},
		Labels: map[string]string{
			labels.SwapperLabel: targetContainerId,
		},
	}

	// Same socket-location convention as internal/patterns/portainer.go.
	dockerSocketHostPath := strings.ReplaceAll(client.DefaultDockerHost, "unix://", "")

	swapperHostConfig := &container.HostConfig{
		Binds: []string{dockerSocketHostPath + ":" + dockerSocketContainerPath},
		// The swapper is a one-shot: never restart it, never auto-remove it
		// (its logs are the only trace of a failed swap).
		AutoRemove: false,
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyDisabled,
		},
	}

	createResponse, err := c.dockerAPI.ContainerCreate(
		ctx,
		swapperConfig,
		swapperHostConfig,
		nil,
		nil,
		selfswap.SwapperContainerName,
	)
	if err != nil {
		return rerrors.Wrap(err, "error creating swapper container")
	}

	err = c.dockerAPI.ContainerStart(ctx, createResponse.ID, container.StartOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error starting swapper container")
	}

	return nil
}
