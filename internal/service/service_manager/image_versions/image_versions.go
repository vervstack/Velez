package image_versions

import (
	"context"
	"slices"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/imagetags"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/user_errors"
)

type ServiceImpl struct {
	runtimes  container_runtime.RuntimeResolver
	dockerAPI client.APIClient
	tags      *imagetags.Client
}

func New(runtimes container_runtime.RuntimeResolver, dockerAPI client.APIClient) *ServiceImpl {
	return &ServiceImpl{runtimes: runtimes, dockerAPI: dockerAPI, tags: imagetags.New()}
}

func (s *ServiceImpl) ListContainerImageVersions(
	ctx context.Context,
	req *velez_api.ListContainerImageVersions_Request,
) (*velez_api.ListContainerImageVersions_Response, error) {
	resp := &velez_api.ListContainerImageVersions_Response{}

	runtime, err := s.runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving runtime")
	}

	info, found, err := runtime.InspectAny(ctx, req.GetContainerId())
	if err != nil {
		return nil, rerrors.Wrap(err, "error inspecting container")
	}

	if !found {
		return nil, rerrors.Wrap(user_errors.ErrRegisterContainerNotFound)
	}

	img, err := s.dockerAPI.ImageInspect(ctx, info.Image)
	if err != nil {
		return nil, rerrors.Wrap(err, "error inspecting image")
	}

	var configImage string

	if info.Config != nil {
		configImage = info.Config.Image
	}

	repository, digest, isResolved := selectRepositoryAndDigest(configImage, img.RepoDigests)
	if !isResolved {
		return resp, nil
	}

	tags, err := s.tags.TagsForDigest(ctx, repository, digest)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing tags for digest")
	}

	slices.Sort(tags)

	resp.Tags = tags

	return resp, nil
}
