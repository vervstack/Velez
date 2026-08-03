package container_manager

import (
	"context"
	"sort"
	"strings"
	"time"

	errors "go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *ContainerManager) InspectSmerd(ctx context.Context, contId string) (*velez_api.Smerd, error) {
	contInfo, err := c.dockerAPI.ContainerInspect(ctx, contId)
	if err != nil {
		return nil, errors.Wrap(err, "error inspecting container")
	}

	// contInfo.Name is the real Docker name, which carries the environment's
	// suffix for a non-empty one (labelBasedRuntime.ContainerCreate). Strip it
	// back to the virtual/logical name using the suffix this same container
	// was stamped with at creation (labels.SuffixLabel) - InspectSmerd bypasses
	// ContainerRuntime entirely (Inspect isn't part of that interface yet, see
	// docs/container_runtimes/roadmap.md), so it can't ask a resolved runtime
	// to do this; reading the suffix straight off the container's own label is
	// equivalent and needs no environment parameter threaded in here.
	var suffix string

	if contInfo.Config != nil {
		suffix = contInfo.Config.Labels[labels.SuffixLabel]
	}

	bareName := strings.Replace(contInfo.Name, "/", "", 1)

	smerd := &velez_api.Smerd{
		Uuid:    contInfo.ID,
		Name:    container_runtime.StripEnvironmentSuffix(bareName, suffix),
		Ports:   parser.ToPortsMapping(contInfo.HostConfig.PortBindings),
		Volumes: parser.ToVolume(contInfo.HostConfig.Mounts),
		Env:     parser.ToDockerEnv(contInfo.Config.Env),
		Labels:  contInfo.Config.Labels,
	}

	imageInfo, err := c.dockerAPI.ImageInspect(ctx, contInfo.Image)
	if err != nil {
		return nil, errors.Wrap(err, "error getting image info")
	}

	for _, imageName := range imageInfo.RepoTags[:1] {
		smerd.ImageName = imageName
	}

	if contInfo.State != nil {
		smerd.Status = velez_api.Smerd_Status(
			velez_api.Smerd_Status_value[contInfo.State.Status])
	}

	createdAt, err := time.Parse("2006-01-02T15:04:05Z", contInfo.Created)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing created at time")
	}

	smerd.CreatedAt = timestamppb.New(createdAt)

	for netName, net := range contInfo.NetworkSettings.Networks {
		nb := &velez_api.NetworkBind{
			NetworkName: netName,
		}

		if len(net.DNSNames) != 0 {
			nb.Aliases = make([]string, 0, len(net.DNSNames)-1)
		}

		for _, dName := range net.DNSNames {
			if !strings.HasPrefix(contInfo.ID, dName) {
				nb.Aliases = append(nb.Aliases, dName)
			}
		}

		smerd.Networks = append(smerd.Networks, nb)
	}

	sort.Slice(smerd.GetNetworks(), func(i, j int) bool {
		return smerd.GetNetworks()[i].GetNetworkName() < smerd.GetNetworks()[j].GetNetworkName()
	})

	return smerd, nil
}
