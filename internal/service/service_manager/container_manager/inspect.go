package container_manager

import (
	"context"
	"time"

	errors "go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils/parser"
	"github.com/godverv/Velez/pkg/velez_api"
)

func (c *ContainerManager) InspectSmerd(ctx context.Context, contId string) (*velez_api.Smerd, error) {
	contInfo, err := c.docker.InspectContainer(ctx, contId)
	if err != nil {
		return nil, errors.Wrap(err, "error inspecting container")
	}

	smerd := &velez_api.Smerd{
		Uuid:    contInfo.ContainerJSONBase.ID,
		Name:    contInfo.ContainerJSONBase.Name,
		Ports:   parser.ToPorts(contInfo.ContainerJSONBase.HostConfig.PortBindings),
		Volumes: parser.ToBind(contInfo.ContainerJSONBase.HostConfig.Mounts),

		Labels: contInfo.Config.Labels,
	}

	imageInfo, err := c.docker.InspectImage(ctx, contInfo.ContainerJSONBase.Image)
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
	return smerd, nil
}
