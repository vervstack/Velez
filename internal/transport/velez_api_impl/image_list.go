package velez_api_impl

import (
	"context"
	"sort"
	"strings"

	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/clients/registryclients"
	"go.vervstack.ru/Velez/internal/domain"
)

// maxGenericV2SearchResults caps how many repositories from a generic_v2
// registry's catalog get a tags/list lookup - the catalog can be large and
// the v2 API has no search endpoint, so every candidate costs one extra
// request.
const (
	maxGenericV2SearchResults = 10
)

func (impl *Impl) SearchImages(ctx context.Context, req *velez_api.SearchImages_Request) (
	*velez_api.SearchImages_Response, error,
) {
	reg, err := impl.resolveSearchRegistry(ctx, req.GetRegistryId())
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving registry")
	}

	var images []*velez_api.SearchImageItem

	switch reg.Type {
	case domain.RegistryTypeGenericV2:
		images, err = impl.searchGenericV2Images(ctx, reg, req)
		if err != nil {
			return nil, rerrors.Wrap(err, "error searching generic_v2 registry")
		}
	default:
		searchReq := domain.ImageSearchRequest{
			Term:            req.GetName(),
			UseOfficialOnly: toolbox.FromPtr(req.UseOnlyOfficial),
		}

		images, err = dockerutils.SearchImages(ctx, impl.dockerAPI, searchReq)
		if err != nil {
			return nil, rerrors.Wrap(err, "error searching images")
		}
	}

	return &velez_api.SearchImages_Response{
		Images: images,
	}, nil
}

// resolveSearchRegistry picks the registry SearchImages should search: the
// one named by id, or - if no id was given - whichever registry is marked
// default. If neither resolves, it falls back to today's behavior (dockerhub,
// unauthenticated): a zero-value domain.Registry with Type
// RegistryTypeDockerHub, which SearchImages's default branch handles the same
// way as an explicit dockerhub registry (it never reads Url/Username/Secret).
func (impl *Impl) resolveSearchRegistry(ctx context.Context, id int64) (domain.Registry, error) {
	if id != 0 {
		reg, err := impl.vervServices.GetRegistry(ctx, id)
		if err != nil {
			return domain.Registry{}, rerrors.Wrap(err, "error getting registry")
		}

		return reg, nil
	}

	registries, err := impl.vervServices.ListRegistries(ctx)
	if err != nil {
		return domain.Registry{}, rerrors.Wrap(err, "error listing registries")
	}

	for _, reg := range registries {
		if reg.IsDefault {
			return reg, nil
		}
	}

	return domain.Registry{Type: domain.RegistryTypeDockerHub}, nil
}

func (impl *Impl) searchGenericV2Images(
	ctx context.Context,
	reg domain.Registry,
	req *velez_api.SearchImages_Request,
) ([]*velez_api.SearchImageItem, error) {
	regClient := registryclients.New(reg)

	repos, err := regClient.Catalog(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing catalog")
	}

	term := strings.ToLower(req.GetName())

	matched := make([]string, 0, len(repos))
	for _, repo := range repos {
		if term == "" || strings.Contains(strings.ToLower(repo), term) {
			matched = append(matched, repo)
		}
	}

	sort.Strings(matched)

	if len(matched) > maxGenericV2SearchResults {
		matched = matched[:maxGenericV2SearchResults]
	}

	out := make([]*velez_api.SearchImageItem, 0, len(matched))
	for _, repo := range matched {
		out = append(out, &velez_api.SearchImageItem{
			Name:      repo,
			LatestTag: firstTagOrEmpty(ctx, regClient, repo),
		})
	}

	return out, nil
}

// firstTagOrEmpty returns the first tag reported for repo, or "" if the
// per-repo tags/list lookup fails - a failure here shouldn't fail the whole
// search. The API gives no ordering guarantee, so "first" is a known
// limitation, not a bug.
func firstTagOrEmpty(ctx context.Context, client *registryclients.Client, repo string) string {
	tags, err := client.TagsList(ctx, repo)
	if err != nil || len(tags) == 0 {
		return ""
	}

	return tags[0]
}
