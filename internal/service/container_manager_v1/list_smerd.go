package container_manager_v1

import (
	"context"

	"github.com/godverv/Velez/pkg/velez_api"
)

const maxList = uint32(10)

func (c *containerManager) ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error) {
	return listContainers(ctx, c.docker, req)
}
