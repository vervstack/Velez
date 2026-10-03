package image_versions

import (
	"regexp"
	"strings"

	"github.com/distribution/reference"
)

var imageIdPattern = regexp.MustCompile(`^(sha256:)?[a-f0-9]{12,64}$`)

type repoDigest struct {
	repository reference.Named
	digest     string
}

// selectRepositoryAndDigest picks the repository the container image was pulled from and the manifest digest
// the local image carries for it. isResolved is false when the image has no registry digest to look up.
func selectRepositoryAndDigest(
	configImage string,
	repoDigests []string,
) (repository reference.Named, digest string, isResolved bool) {
	parsed := parseRepoDigests(repoDigests)
	if len(parsed) == 0 {
		return nil, "", false
	}

	repository = repositoryFromConfigImage(configImage)
	if repository == nil {
		return parsed[0].repository, parsed[0].digest, true
	}

	for _, candidate := range parsed {
		if candidate.repository.Name() == repository.Name() {
			return repository, candidate.digest, true
		}
	}

	return nil, "", false
}

func repositoryFromConfigImage(configImage string) reference.Named {
	if imageIdPattern.MatchString(configImage) {
		return nil
	}

	named, err := reference.ParseNormalizedNamed(configImage)
	if err != nil {
		return nil
	}

	return reference.TrimNamed(named)
}

func parseRepoDigests(repoDigests []string) []repoDigest {
	out := make([]repoDigest, 0, len(repoDigests))

	for _, entry := range repoDigests {
		repoPart, digest, found := strings.Cut(entry, "@")
		if !found {
			continue
		}

		named, err := reference.ParseNormalizedNamed(repoPart)
		if err != nil {
			continue
		}

		out = append(out, repoDigest{repository: reference.TrimNamed(named), digest: digest})
	}

	return out
}
