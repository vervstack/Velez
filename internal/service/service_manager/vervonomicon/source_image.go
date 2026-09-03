package vervonomicon

import (
	"context"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
)

const (
	// vervDirPath is where a descriptor is baked into an image, per
	// docs/features/vervonomicon.md's "Where the descriptor comes from".
	vervDirPath = "/verv"

	// scratchContainerSuffix names the throwaway container ImageSource
	// creates to read .verv/ out of an image. Distinct from
	// internal/jobs.scratchContainerSuffix ("_config_scanning") so the two
	// scratch containers for the same service never collide.
	scratchContainerSuffix = "_verv_scanning"
)

// ImageSource reads .verv/ out of an image through a throwaway, never-started
// scratch container - the same dance internal/jobs/assemble_config.go uses to
// read /app/config/config.yaml (prepareScratchImageJob ->
// createScratchContainerJob -> read -> dropScratchContainerJob), collapsed
// into one synchronous call since this has no need for the jobs engine's
// checkpointing: it's a short read used by the GetVervonomicon RPC and, later,
// the deploy path.
type ImageSource struct {
	nodeClients node_clients.NodeClients
	runtimes    container_runtime.RuntimeResolver
}

func NewImageSource(nodeClients node_clients.NodeClients, runtimes container_runtime.RuntimeResolver) *ImageSource {
	return &ImageSource{
		nodeClients: nodeClients,
		runtimes:    runtimes,
	}
}

// Read ensures imageName is pulled, creates a scratch container from it,
// reads .verv/ out of it, and removes the container on every path, including
// error. An image with no /verv directory returns ErrNoDescriptor, not an
// error - the caller can tell "no descriptor" apart from "read failed".
func (s *ImageSource) Read(ctx context.Context, serviceName, imageName string) (map[string][]byte, error) {
	runtime, err := s.runtimes.Runtime(ctx, "")
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving container runtime")
	}

	_, err = runtime.PullImage(ctx, imageName)
	if err != nil {
		return nil, rerrors.Wrap(err, "error pulling image")
	}

	containerName := serviceName + scratchContainerSuffix

	cfg := &container.Config{
		Image:    imageName,
		Hostname: containerName,
	}
	hostCfg := &container.HostConfig{}
	netCfg := &network.NetworkingConfig{}
	platform := &ocispec.Platform{}

	created, err := s.nodeClients.Docker().ContainerCreate(
		// Throwaway, never-started scratch container - not owned by any
		// environment.
		ctx, cfg, hostCfg, netCfg, platform, containerName, "")
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating scratch container")
	}

	defer s.dropScratchContainer(ctx, created.ID)

	files, err := dockerutils.ReadDirFromContainer(ctx, s.nodeClients.Docker().Client(), created.ID, vervDirPath)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, ErrNoDescriptor
		}

		return nil, rerrors.Wrap(err, "error reading .verv from image")
	}

	return files, nil
}

func (s *ImageSource) dropScratchContainer(ctx context.Context, containerID string) {
	err := s.nodeClients.Docker().Remove(ctx, containerID)
	if err != nil && !errdefs.IsNotFound(err) {
		log.Ctx(ctx).Warn().
			Str("container_id", containerID).
			Err(err).
			Msg("error dropping verv scratch container")
	}
}
