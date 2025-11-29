package dockerutils

import (
	"context"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	errors "go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func ListImages(ctx context.Context, docker client.APIClient, req domain.ImageListRequest) ([]*velez_api.Image, error) {
	dockerReq := image.ListOptions{
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

func SearchImages(ctx context.Context, docker client.APIClient, req domain.ImageSearchRequest) ([]any, error) {
	listReq := registry.SearchOptions{
		RegistryAuth:  "",
		PrivilegeFunc: nil,
		Filters:       filters.NewArgs(),
		Limit:         0,
	}

	if req.UseOfficialOnly {
		listReq.Filters.Add("is-official", "true")
	}

	images, err := docker.ImageSearch(ctx, req.Term, listReq)
	if err != nil {
		return nil, errors.Wrap(err, "error searching images")
	}

	_ = images

	return nil, nil

}
