package dockerutils

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/pkg/velez_api"
)

func ListImages(ctx context.Context, docker client.CommonAPIClient, req domain.ImageListRequest) ([]*velez_api.Image, error) {
	dockerReq := types.ImageListOptions{
		Filters: filters.NewArgs(),
	}
	dockerReq.Filters.Add("reference", req.Name)

	images, err := docker.ImageList(ctx, dockerReq)
	if err != nil {
		return nil, errors.Wrap(err, "error listing images")
	}

	resp := make([]*velez_api.Image, len(images))
	for i := range images {
		if images[i].RepoTags[0] == "" {
			continue
		}

		resp[i] = &velez_api.Image{
			Name: images[i].RepoTags[0],
			Tags: images[i].RepoTags,
		}
	}

	return resp, nil
}
