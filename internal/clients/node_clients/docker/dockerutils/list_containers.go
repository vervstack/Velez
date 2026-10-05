package dockerutils

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	errors "go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/list_request"
)

// ListContainers returns every matching container unless req carries an
// explicit limit. Internal node-wide callers (local_storage's label-derived
// views, environment cascades, upgrades) must never see a silently truncated
// list - a dind/runner/service older than the newest N containers on a busy
// daemon would otherwise vanish from them. Paging for the public ListSmerds
// API lives in container_manager.ListSmerds.
func ListContainers(
	ctx context.Context,
	docker client.APIClient,
	req *velez_api.ListSmerds_Request,
) ([]container.Summary, error) {
	dockerReq := container.ListOptions{
		All: true,
	}

	filter := list_request.New()

	if req.GetLimit() != 0 {
		dockerReq.Limit = int(req.GetLimit())
	}

	if req.GetId() != "" {
		filter.Id(req.GetId())
	}

	if req.GetName() != "" {
		filter.Name(req.GetName())
	}

	if req.GetLabel() != nil {
		for k, v := range req.GetLabel() {
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
