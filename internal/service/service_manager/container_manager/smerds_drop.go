package container_manager

import (
	"context"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (c *ContainerManager) DropSmerds(ctx context.Context, req *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error) {
	out := &velez_api.DropSmerd_Response{}

	for _, uuid := range append(req.Uuids, req.Name...) {
		err := c.docker.Remove(ctx, uuid)
		if err == nil {
			out.Successful = append(out.Successful, uuid)
			continue
		}

		out.Failed = append(out.Failed,
			&velez_api.DropSmerd_Response_Error{
				Uuid:  uuid,
				Cause: err.Error(),
			})
	}

	return out, nil
}
