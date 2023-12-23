package v1

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"

	"github.com/godverv/Velez/internal/domain"
)

func listImages(ctx context.Context, docker client.CommonAPIClient, req domain.ImageListRequest) ([]domain.Image, error) {
	dockerReq := types.ImageListOptions{
		Filters: filters.NewArgs(),
	}
	dockerReq.Filters.Add("reference", req.ImageName)

	images, err := docker.ImageList(ctx, dockerReq)
	if err != nil {
		return nil, errors.Wrap(err, "error listing images")
	}

	resp := make([]domain.Image, len(images))
	for i := range images {
		if images[i].RepoTags[0] == "" {
			continue
		}

		resp[i] = domain.Image{
			Name: images[i].RepoTags[0],
			Tags: images[i].RepoTags,
		}
	}

	return resp, nil
}

func pullImage(ctx context.Context, docker client.CommonAPIClient, req domain.Image) (domain.Image, error) {
	rdr, err := docker.ImagePull(ctx, req.Name, types.ImagePullOptions{})
	if err != nil {
		return domain.Image{}, errors.Wrap(err, "error pulling image")
	}
	_, err = io.ReadAll(rdr)
	if err != nil {
		return domain.Image{}, errors.Wrap(err, "error reading pull log")
	}

	err = rdr.Close()
	if err != nil {
		return domain.Image{}, errors.Wrap(err, "error closing image pull reader")
	}

	dockerReq := types.ImageListOptions{
		Filters: filters.NewArgs(),
	}

	dockerReq.Filters.Add("reference", req.Name)

	imageList, err := docker.ImageList(ctx, dockerReq)
	if err != nil {
		return domain.Image{}, errors.Wrap(err, "error listing images after pulling")
	}

	if len(imageList) == 0 {
		// TODO - is it a possible situation though?
		return domain.Image{}, errors.New("image list is empty")
	}

	if len(imageList[0].RepoTags) == 0 {
		// TODO - is it a possible situation though?
		return domain.Image{}, errors.New("image has no tags")
	}

	return domain.Image{
		Name: imageList[0].RepoTags[0],
		Tags: imageList[0].RepoTags,
	}, nil
}
