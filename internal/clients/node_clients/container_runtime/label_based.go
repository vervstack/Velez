package container_runtime

import (
	"context"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	labelValueTrue = "true"

	// nameSuffixSeparator joins a smerd's logical name and its environment
	// suffix. It reproduces, byte for byte, the convention the pre-environments
	// pipeliner used (internal/pipelines/do_smerd_launch.go's
	// `req.Name = req.GetName() + "_" + p.suffix`).
	nameSuffixSeparator = "_"
)

// labelBasedRuntime is tier 1, the default: every environment on the node
// shares one Docker daemon and is kept apart by
//
//   - the labels.SuffixLabel stamped on each container (what ListSmerds and
//     friends filter on), and
//   - the actual Docker container NAME, which carries the same suffix so two
//     environments deploying the same logical smerd name don't collide on the
//     daemon's globally-unique container namespace.
//
// An empty suffix - today's default/PROD on a node that never configured
// ContainerSuffix - means "unsuffixed", preserving pre-multi-environment
// container naming exactly.
type labelBasedRuntime struct {
	commonRuntime

	suffix string
	// bakedLabels are the node's configured CustomLabels, in "name=value" (or
	// bare "name") form, stamped onto every container Velez creates - same as
	// docker.Docker's own bakedLabels.
	bakedLabels []string
}

// ContainerCreate reimplements docker.Docker.ContainerCreate's label stamping
// and conflict handling, and adds the name-conflict resolution that stamping
// alone never provided.
func (r *labelBasedRuntime) ContainerCreate(
	ctx context.Context,
	req ContainerCreateRequest,
) (container.CreateResponse, error) {
	if req.Config == nil || req.Config.Config == nil {
		return container.CreateResponse{}, rerrors.New("container config is required")
	}

	config := req.Config.Config

	if config.Labels == nil {
		config.Labels = map[string]string{}
	}

	config.Labels[labels.CreatedWithVelezLabel] = labelValueTrue
	// The suffix is the resolved environment's, decided when this runtime was
	// handed out. Empty means "the node's unscoped/default environment",
	// preserving pre-multi-environment behavior.
	config.Labels[labels.SuffixLabel] = r.suffix

	for _, label := range r.bakedLabels {
		before, after, ok := strings.Cut(label, "=")
		name := label

		var val string

		if ok {
			val = after
			name = before
		}

		config.Labels[name] = val
	}

	var hostConfig *container.HostConfig

	if req.HostConfig != nil {
		hostConfig = req.HostConfig.HostConfig
	}

	var networkingConfig *network.NetworkingConfig

	if req.NetworkingConfig != nil {
		networkingConfig = req.NetworkingConfig.NetworkingConfig
	}

	createResponse, err := r.cli.ContainerCreate(
		ctx, config, hostConfig, networkingConfig, req.Platform, r.containerName(req.ContainerName))
	if err != nil {
		if errdefs.IsConflict(err) {
			return container.CreateResponse{}, rerrors.Wrap(docker.HandleConflictMessage(err))
		}

		return container.CreateResponse{}, rerrors.Wrap(err, "error during container creation via docker api")
	}

	return createResponse, nil
}

// ListContainers scopes results to r.suffix - the environment this runtime
// instance was resolved for - the same way docker.Docker.ListContainers does:
// an empty suffix means "don't scope by environment" (today's default/PROD),
// a non-empty suffix filters on labels.SuffixLabel.
func (r *labelBasedRuntime) ListContainers(
	ctx context.Context,
	req *velez_api.ListSmerds_Request,
) ([]container.Summary, error) {
	if req.Label == nil {
		req.Label = map[string]string{}
	}

	if r.suffix != "" {
		req.Label[labels.SuffixLabel] = r.suffix
	}

	list, err := dockerutils.ListContainers(ctx, r.cli, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	return list, nil
}

// containerName resolves the logical smerd name into the actual Docker
// container name for this environment: the bare name when the suffix is empty,
// "<name>_<suffix>" otherwise.
func (r *labelBasedRuntime) containerName(name string) string {
	if r.suffix == "" {
		return name
	}

	return name + nameSuffixSeparator + r.suffix
}
