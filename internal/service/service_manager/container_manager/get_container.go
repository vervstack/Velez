package container_manager

import (
	"context"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"

	errors "go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/cluster/env/containerinfo"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetContainer inspects a single container on the daemon backing
// req.Environment by literal Docker container id, registered with Velez or
// not - unlike InspectSmerd, there is no ownership check, matching
// ContainerRuntime.InspectAny (see its doc comment for why that's unscoped
// by design).
func (c *ContainerManager) GetContainer(
	ctx context.Context,
	req *velez_api.GetContainer_Request,
) (*velez_api.DockerContainer, error) {
	runtime, err := c.runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return nil, errors.Wrap(err, "error resolving environment")
	}

	contInfo, found, err := runtime.InspectAny(ctx, req.GetId())
	if err != nil {
		return nil, errors.Wrap(err, "error inspecting container")
	}

	if !found {
		return nil, errors.New("container %q not found", req.GetId())
	}

	dc := &velez_api.DockerContainer{
		Id:           contInfo.ID,
		Name:         strings.Replace(contInfo.Name, "/", "", 1),
		Labels:       contInfo.Config.Labels,
		Env:          parser.ToDockerEnv(contInfo.Config.Env),
		Mounts:       toMounts(contInfo.Mounts),
		IsRegistered: contInfo.Config.Labels[labels.CreatedWithVelezLabel] == labelTrue,

		SuggestedPattern: suggestedPattern(contInfo.Config.Image),
	}

	if containerinfo.IsSelf(contInfo.ID) {
		dc.IsRegistered = true
	}

	dc.SuggestedRunnerDefaults = suggestedRunnerDefaults(contInfo.Config.Image, dc.GetLabels(), dc.GetEnv())

	if contInfo.NetworkSettings != nil {
		dc.Networks = toNetworkBinds(contInfo.ID, contInfo.NetworkSettings.Networks)
	}

	ownerId, isShared, err := resolveNetworkOwnerId(ctx, runtime, contInfo)
	if err != nil {
		return nil, errors.Wrap(err, "error resolving network owner")
	}

	if isShared {
		dc.NetworkOwnerContainerId = &ownerId
	}

	svc, isLabelled := contInfo.Config.Labels[labels.VervServiceLabel]
	if isLabelled {
		dc.LinkedServiceName = &svc
	}

	if !isLabelled {
		boundServices, bindErr := c.boundServiceNames(ctx, req.GetEnvironment())
		if bindErr != nil {
			return nil, errors.Wrap(bindErr, "error resolving container bindings")
		}

		boundSvc, isBound := boundServices[dc.GetName()]
		if isBound {
			dc.LinkedServiceName = &boundSvc
			dc.IsRegistered = true
		}
	}

	imageInfo, err := c.dockerAPI.ImageInspect(ctx, contInfo.Image)
	if err != nil {
		return nil, errors.Wrap(err, "error getting image info")
	}

	for _, imageName := range imageInfo.RepoTags[:1] {
		dc.ImageName = imageName
	}

	if contInfo.State != nil {
		dc.Status = velez_api.Smerd_Status(velez_api.Smerd_Status_value[contInfo.State.Status])
	}

	createdAt, err := time.Parse("2006-01-02T15:04:05Z", contInfo.Created)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing created at time")
	}

	dc.CreatedAt = timestamppb.New(createdAt)

	return dc, nil
}

const (
	containerNetworkModePrefix = "container:"
	maxNetworkOwnerHops        = 5
)

func resolveNetworkOwnerId(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	contInfo container.InspectResponse,
) (string, bool, error) {
	current := contInfo
	isShared := false

	for range maxNetworkOwnerHops {
		if current.HostConfig == nil {
			break
		}

		target, hasOwner := strings.CutPrefix(string(current.HostConfig.NetworkMode), containerNetworkModePrefix)
		if !hasOwner {
			break
		}

		owner, found, err := runtime.InspectAny(ctx, target)
		if err != nil {
			return "", false, errors.Wrap(err, "error inspecting network owner")
		}

		if !found {
			break
		}

		current = owner
		isShared = true
	}

	return current.ID, isShared, nil
}
