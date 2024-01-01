package container_manager_v1

import (
	"context"

	"github.com/docker/docker/api/types"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (c *containerManager) DropSmerds(ctx context.Context, req *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error) {
	out := &velez_api.DropSmerd_Response{}

	for _, arg := range append(req.Uuids, req.Name...) {
		err := c.docker.ContainerRemove(ctx, arg, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
		if err != nil {
			out.Failed = append(out.Failed, &velez_api.DropSmerd_Response_Error{
				Uuid:  arg,
				Cause: err.Error(),
			})
		} else {
			out.Successful = append(out.Successful, arg)
		}
	}

	return out, nil
}
