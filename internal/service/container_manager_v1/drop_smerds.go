package container_manager_v1

import (
	"context"

	"github.com/docker/docker/api/types"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (c *containerManager) DropSmerds(ctx context.Context, req *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error) {
	out := &velez_api.DropSmerd_Response{}
	for _, uuid := range req.Uuids {
		err := c.docker.ContainerRemove(ctx, uuid, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
		if err != nil {
			out.Failed = append(out.Failed, &velez_api.DropSmerd_Response_Error{
				Uuid:  uuid,
				Cause: err.Error(),
			})
		} else {
			out.Successful = append(out.Successful, uuid)
		}
	}

	return out, nil
}
