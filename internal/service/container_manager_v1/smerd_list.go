package container_manager_v1

import (
	"context"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/pkg/velez_api"
)

func (c *containerManager) ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error) {
	return dockerutils.ListContainers(ctx, c.docker, req)
}
