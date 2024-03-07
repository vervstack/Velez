package container_manager_v1

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils/parser"
	"github.com/godverv/Velez/pkg/velez_api"
)

func (c *ContainerManager) ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error) {
	cl, err := dockerutils.ListContainers(ctx, c.docker, req)
	if err != nil {
		return nil, errors.Wrap(err, "error listing containers")
	}

	resp := &velez_api.ListSmerds_Response{
		Smerds: make([]*velez_api.Smerd, len(cl)),
	}

	for i, item := range cl {
		resp.Smerds[i] = &velez_api.Smerd{
			Uuid:      item.ID,
			ImageName: item.Image,

			Status: velez_api.Smerd_Status(velez_api.Smerd_Status_value[item.State]),
			CreatedAt: &timestamppb.Timestamp{
				Seconds: item.Created,
			},

			Ports:   parser.ToPorts(item.Ports),
			Volumes: parser.ToVolumes(item.Mounts),
		}

		if len(item.Names) != 0 {
			resp.Smerds[i].Name = item.Names[0][1:]
		}
	}

	return resp, nil
}
