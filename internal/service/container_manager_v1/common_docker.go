package container_manager_v1

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/internal/utils/comparator"
	"github.com/godverv/Velez/pkg/velez_api"
)

func listImages(ctx context.Context, docker client.CommonAPIClient, req domain.ImageListRequest) ([]*velez_api.Image, error) {
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

func pullImage(ctx context.Context, docker client.CommonAPIClient, req domain.ImageListRequest) (*velez_api.Image, error) {
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

func listContainers(ctx context.Context, docker client.CommonAPIClient, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error) {
	dockerReq := types.ContainerListOptions{
		All:     true,
		Filters: filters.NewArgs(),
	}

	if req.Limit != nil {
		dockerReq.Limit = int(comparator.LessInt(maxList, *req.Limit))
	} else {
		dockerReq.Limit = int(maxList)
	}

	cl, err := docker.ContainerList(ctx, dockerReq)
	if err != nil {
		return nil, errors.Wrap(err, "error listing containers")
	}

	resp := &velez_api.ListSmerds_Response{
		Smerds: make([]*velez_api.Smerd, len(cl)),
	}

	for i, item := range cl {
		resp.Smerds[i] = &velez_api.Smerd{
			Uuid:      item.ID,
			ImageName: item.Image,

			Status: velez_api.Smerd_Status(velez_api.Smerd_Status_value[item.State]),
			CreatedAt: &timestamppb.Timestamp{
				Seconds: item.Created,
			},

			Ports:   toPorts(item.Ports),
			Volumes: toVolumes(item.Mounts),
		}

		if len(item.Names) != 0 {
			resp.Smerds[i].Name = item.Names[0][1:]
		}

	}

	return resp, nil
}
