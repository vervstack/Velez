package dockerutils

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils/list_request"
	"github.com/godverv/Velez/internal/utils/common"
	"github.com/godverv/Velez/pkg/velez_api"
)

const maxList = 10

func ListContainers(ctx context.Context, docker client.CommonAPIClient, req *velez_api.ListSmerds_Request) ([]types.Container, error) {
	dockerReq := types.ContainerListOptions{
		All:   true,
		Limit: maxList,
	}

	filter := list_request.New()

	if req.Limit != nil {
		dockerReq.Limit = common.Less[int](maxList, int(req.GetLimit()))
	}

	if req.GeneralSearch != nil {
		filter.Id(req.GetGeneralSearch())
		filter.Name(req.GetGeneralSearch())
		filter.Label(req.GetGeneralSearch())
	}

	if req.Id != nil {
		filter.Id(req.GetId())
	}

	if req.Name != nil {
		filter.Name(req.GetName())
	}

	dockerReq.Filters = filter.Args()

	cl, err := docker.ContainerList(ctx, dockerReq)
	if err != nil {
		return nil, errors.Wrap(err, "error listing containers")
	}

	return cl, nil
}
