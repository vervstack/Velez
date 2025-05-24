package dockerutils

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

func PullImage(ctx context.Context, docker client.CommonAPIClient, name string, force bool) (_ *velez_api.Image, err error) {
	var imageList []image.Summary

	dockerReq := image.ListOptions{
		Filters: filters.NewArgs(),
	}
	dockerReq.Filters.Add("reference", name)

	if !force {
		imageList, err = docker.ImageList(ctx, dockerReq)
		if err != nil {
			return nil, errors.Wrap(err, "error listing images after pulling")
		}
	}

	if len(imageList) == 0 {
		var rdr io.ReadCloser
		rdr, err = docker.ImagePull(ctx, name, image.PullOptions{})
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

		imageList, err = docker.ImageList(ctx, dockerReq)
		if err != nil {
			return nil, errors.Wrap(err, "error listing images after pulling")
		}

	}

	if len(imageList) == 0 {
		return nil, errors.New("image list is empty")
	}

	if len(imageList[0].RepoTags) == 0 {
		return nil, errors.New("image has no tags")
	}

	return &velez_api.Image{
		Name:   imageList[0].RepoTags[0],
		Tags:   imageList[0].RepoTags,
		Labels: imageList[0].Labels,
	}, nil
}
