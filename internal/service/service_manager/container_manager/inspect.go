package container_manager

import (
	"context"
	"sort"
	"strings"
	"time"

	errors "go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.vervstack.ru/Velez/internal/clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (c *ContainerManager) InspectSmerd(ctx context.Context, contId string) (*velez_api.Smerd, error) {
	contInfo, err := c.dockerAPI.ContainerInspect(ctx, contId)
	if err != nil {
		return nil, errors.Wrap(err, "error inspecting container")
	}

	smerd := &velez_api.Smerd{
		Uuid:    contInfo.ContainerJSONBase.ID,
		Name:    contInfo.ContainerJSONBase.Name,
		Ports:   parser.ToPortsMapping(contInfo.ContainerJSONBase.HostConfig.PortBindings),
		Volumes: parser.ToVolume(contInfo.ContainerJSONBase.HostConfig.Mounts),
		Env:     parser.ToDockerEnv(contInfo.Config.Env),
		Labels:  contInfo.Config.Labels,
	}

	imageInfo, err := c.dockerAPI.ImageInspect(ctx, contInfo.ContainerJSONBase.Image)
	if err != nil {
		return nil, errors.Wrap(err, "error getting image info")
	}

	for _, imageName := range imageInfo.RepoTags[:1] {
		smerd.ImageName = imageName
	}

	if contInfo.ContainerJSONBase.State != nil {
		smerd.Status = velez_api.Smerd_Status(
			velez_api.Smerd_Status_value[contInfo.ContainerJSONBase.State.Status])
	}

	createdAt, err := time.Parse("2006-01-02T15:04:05Z", contInfo.ContainerJSONBase.Created)
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

	sort.Slice(smerd.Networks, func(i, j int) bool {
		return smerd.Networks[i].NetworkName < smerd.Networks[j].NetworkName
	})

	return smerd, nil
}
