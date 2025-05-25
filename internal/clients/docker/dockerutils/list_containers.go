package dockerutils

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	errors "go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/docker/dockerutils/list_request"
	"go.vervstack.ru/Velez/internal/utils/common"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

const maxList = 10

func ListContainers(ctx context.Context, docker client.APIClient, req *velez_api.ListSmerds_Request) ([]types.Container, error) {
	dockerReq := container.ListOptions{
		All:   true,
		Limit: maxList,
	}

	filter := list_request.New()

	if req.GetLimit() != 0 {
		dockerReq.Limit = common.Less[int](maxList, int(req.GetLimit()))
	}

	if req.GetId() != "" {
		filter.Id(req.GetId())
	}

	if req.GetName() != "" {
		filter.Name(req.GetName())
	}

	if req.Label != nil {
		for k, v := range req.Label {
			filter.Label(k + "=" + v)
		}

	}

	dockerReq.Filters = filter.Args()

	cl, err := docker.ContainerList(ctx, dockerReq)
	if err != nil {
		return nil, errors.Wrap(err, "error listing containers")
	}

	return cl, nil
}
