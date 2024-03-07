package dockerutils

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/pkg/velez_api"
)

func PullImage(ctx context.Context, docker client.CommonAPIClient, req domain.ImageListRequest) (*velez_api.Image, error) {
	rdr, err := docker.ImagePull(ctx, req.Name, types.ImagePullOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error pulling image")
	}

	_, err = io.ReadAll(rdr)
	if err != nil {
		return nil, errors.Wrap(err, "error reading pull log")
	}

	err = rdr.Close()
	if err != nil {
		return nil, errors.Wrap(err, "error closing image pull reader")
	}

	dockerReq := types.ImageListOptions{
		Filters: filters.NewArgs(),
	}

	dockerReq.Filters.Add("reference", req.Name)

	imageList, err := docker.ImageList(ctx, dockerReq)
	if err != nil {
		return nil, errors.Wrap(err, "error listing images after pulling")
	}

	if len(imageList) == 0 {
		return nil, errors.New("image list is empty")
	}

	if len(imageList[0].RepoTags) == 0 {
		return nil, errors.New("image has no tags")
	}

	return &velez_api.Image{
		Name: imageList[0].RepoTags[0],
		Tags: imageList[0].RepoTags,
	}, nil
}
