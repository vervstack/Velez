package container_manager

import (
	"context"

	errors "go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/godverv/Velez/internal/domain/labels"
	"github.com/godverv/Velez/pkg/velez_api"
)

func (c *ContainerManager) ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error) {
	cl, err := c.docker.ListContainers(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "error listing containers")
	}

	resp := &velez_api.ListSmerds_Response{
		Smerds: make([]*velez_api.Smerd, 0, len(cl)),
	}

	for _, container := range cl {
		if container.Labels[labels.CreatedWithVelezLabel] != "true" {
			continue
		}

		smerd := &velez_api.Smerd{
			Uuid:      container.ID,
			ImageName: container.Image,

			Status: velez_api.Smerd_Status(velez_api.Smerd_Status_value[container.State]),
			CreatedAt: &timestamppb.Timestamp{
				Seconds: container.Created,
			},

			Labels: container.Labels,
		}

		if len(container.Names) != 0 {
			smerd.Name = container.Names[0][1:]
		}
		resp.Smerds = append(resp.Smerds, smerd)
	}

	return resp, nil
}
