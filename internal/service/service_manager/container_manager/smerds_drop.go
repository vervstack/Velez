package container_manager

import (
	"context"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (c *ContainerManager) DropSmerds(
	ctx context.Context,
	req *velez_api.DropSmerd_Request,
) (*velez_api.DropSmerd_Response, error) {
	out := &velez_api.DropSmerd_Response{}

	worklist := make([]string, 0, len(req.GetUuids())+len(req.GetName()))

	worklist = append(worklist, req.GetUuids()...)
	worklist = append(worklist, req.GetName()...)

	runtime, err := c.runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		for _, identifier := range worklist {
			out.Failed = append(out.Failed,
				&velez_api.DropSmerd_Response_Error{
					Uuid:  identifier,
					Cause: err.Error(),
				})
		}

		return out, nil
	}

	for _, identifier := range worklist {
		err = runtime.Remove(ctx, identifier)
		if err == nil {
			out.Successful = append(out.Successful, identifier)

			continue
		}

		out.Failed = append(out.Failed,
			&velez_api.DropSmerd_Response_Error{
				Uuid:  identifier,
				Cause: err.Error(),
			})
	}

	return out, nil
}
