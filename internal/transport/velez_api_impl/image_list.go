package velez_api_impl

import (
	"context"
	"net/url"
	"sort"
	"strings"

	"github.com/distribution/reference"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/clients/registryclients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// maxGenericV2SearchResults caps how many repositories from a generic_v2
// registry's catalog get a tags/list lookup - the catalog can be large and
// the v2 API has no search endpoint, so every candidate costs one extra
// request.
const (
	maxGenericV2SearchResults = 10
)

// dockerHubDomains are the reference domains that mean "Docker Hub" - a name
// carrying one of these is treated as domainless and falls through to the
// id/default/dockerhub resolution.
var dockerHubDomains = map[string]struct{}{
	"docker.io":            {},
	"index.docker.io":      {},
	"registry-1.docker.io": {},
}

func (impl *Impl) SearchImages(ctx context.Context, req *velez_api.SearchImages_Request) (
	*velez_api.SearchImages_Response, error,
) {
	reg, term, err := impl.resolveSearchRegistry(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving registry")
	}

	var images []*velez_api.SearchImageItem

	switch reg.Type {
	case domain.RegistryTypeGenericV2:
		images, err = impl.searchGenericV2Images(ctx, reg, term)
		if err != nil {
			return nil, rerrors.Wrap(err, "error searching generic_v2 registry")
		}
	default:
		searchReq := domain.ImageSearchRequest{
			Term:            term,
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

// resolveSearchRegistry picks the registry SearchImages should search and the
// catalog/search filter term to use against it. Precedence:
//
//  1. an explicit registry_id - statefull mode only, since single-node's
//     registries list is just the fixed builtin entry;
//  2. a registry domain parsed off the front of the search term (Docker's
//     standard image-reference rule), matched against a configured registry
//     or, failing that, searched ad hoc and unauthenticated;
//  3. whichever registry is marked default;
//  4. unauthenticated Docker Hub (a zero-value Registry with Type
//     RegistryTypeDockerHub, handled the same as an explicit dockerhub row).
//
// The returned term equals req.GetName() everywhere except the domain-parsed
// path, where the domain prefix is stripped so searchGenericV2Images's
// substring match still lands.
func (impl *Impl) resolveSearchRegistry(ctx context.Context, req *velez_api.SearchImages_Request) (
	domain.Registry, string, error,
) {
	name := req.GetName()

	if req.GetRegistryId() != 0 {
		if !impl.vervServices.IsStatefull() {
			return domain.Registry{}, "", rerrors.Wrap(user_errors.ErrRequiresStatefullMode)
		}

		reg, err := impl.vervServices.GetRegistry(ctx, req.GetRegistryId())
		if err != nil {
			return domain.Registry{}, "", rerrors.Wrap(err, "error getting registry")
		}

		return reg, name, nil
	}

	regDomain, filterTerm := registryDomainFromRef(name)
	if regDomain != "" {
		reg, err := impl.resolveRegistryByDomain(ctx, regDomain)
		if err != nil {
			return domain.Registry{}, "", rerrors.Wrap(err, "error resolving registry by domain")
		}

		return reg, filterTerm, nil
	}

	registries, err := impl.vervServices.ListRegistries(ctx)
	if err != nil {
		return domain.Registry{}, "", rerrors.Wrap(err, "error listing registries")
	}

	for _, reg := range registries {
		if reg.IsDefault {
			return reg, name, nil
		}
	}

	dockerHub := domain.Registry{Type: domain.RegistryTypeDockerHub}

	return dockerHub, name, nil
}

// resolveRegistryByDomain matches regDomain against a configured registry's
// Url host and returns that registry (so its credentials and type apply). If
// nothing matches it returns an ad-hoc, unauthenticated generic_v2 registry
// pointed straight at https://<regDomain> - a registry the user never
// configured is still searchable, it just gets no auth.
func (impl *Impl) resolveRegistryByDomain(ctx context.Context, regDomain string) (domain.Registry, error) {
	registries, err := impl.vervServices.ListRegistries(ctx)
	if err != nil {
		return domain.Registry{}, rerrors.Wrap(err, "error listing registries")
	}

	for _, reg := range registries {
		host := registryUrlHost(reg.Url)
		if host != "" && strings.EqualFold(host, regDomain) {
			return reg, nil
		}
	}

	adHoc := domain.Registry{
		Type: domain.RegistryTypeGenericV2,
		Url:  "https://" + regDomain,
	}

	return adHoc, nil
}

// registryDomainFromRef parses name as a Docker image reference and returns
// the registry domain to resolve against plus the reference path with that
// domain stripped. It returns "", "" when name carries no domain or a Docker
// Hub one - the standard rule being that the segment before the first "/" is
// a domain only if it holds a "." or ":" or is exactly "localhost".
func registryDomainFromRef(name string) (regDomain, path string) {
	ref, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return "", ""
	}

	d := reference.Domain(ref)

	_, isDockerHub := dockerHubDomains[d]
	if isDockerHub {
		return "", ""
	}

	return d, reference.Path(ref)
}

// registryUrlHost extracts the host[:port] from a stored registry Url,
// tolerating a missing scheme (some rows store a bare "myreg.io:5000").
func registryUrlHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if !strings.Contains(raw, "://") {
		raw = "//" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	return u.Host
}

func (impl *Impl) searchGenericV2Images(
	ctx context.Context,
	reg domain.Registry,
	term string,
) ([]*velez_api.SearchImageItem, error) {
	regClient := registryclients.New(reg)

	repos, err := regClient.Catalog(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing catalog")
	}

	term = strings.ToLower(term)

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
