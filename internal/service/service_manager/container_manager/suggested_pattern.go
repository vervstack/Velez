package container_manager

import (
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon/builtin"
)

// registerPatternDescriptors maps a builtin descriptor to the register
// pattern its image is suggested for. registry_ui is not a pattern.
var (
	registerPatternDescriptors = map[string]velez_api.ServicePattern{
		"postgres":      velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES,
		"registry":      velez_api.ServicePattern_SERVICE_PATTERN_REGISTRY,
		"github_runner": velez_api.ServicePattern_SERVICE_PATTERN_GITHUB_RUNNER,
		"gitlab_runner": velez_api.ServicePattern_SERVICE_PATTERN_GITLAB_RUNNER,
	}

	patternRepositories = sync.OnceValue(loadPatternRepositories)
)

// imageRepositoryName is the last path segment of an image reference without
// registry host, tag or digest: "my.registry:5000/team/postgres:16" -> "postgres".
func imageRepositoryName(image string) string {
	withoutDigest, _, _ := strings.Cut(image, "@")

	lastSegment := withoutDigest[strings.LastIndex(withoutDigest, "/")+1:]

	name, _, _ := strings.Cut(lastSegment, ":")

	return strings.ToLower(name)
}

func loadPatternRepositories() map[string]velez_api.ServicePattern {
	repositories := make(map[string]velez_api.ServicePattern, len(registerPatternDescriptors))

	for descriptorName, pattern := range registerPatternDescriptors {
		files, err := builtin.Read(descriptorName)
		if err != nil {
			log.Error().Err(rerrors.Wrap(err)).Str("descriptor", descriptorName).Msg("error reading builtin descriptor")

			continue
		}

		descriptor, err := vervonomicon.Parse(files)
		if err != nil {
			log.Error().Err(rerrors.Wrap(err)).Str("descriptor", descriptorName).Msg("error parsing builtin descriptor")

			continue
		}

		repository := imageRepositoryName(descriptor.Deployment.App.Image)
		if repository == "" {
			continue
		}

		repositories[repository] = pattern
	}

	return repositories
}

func suggestedPattern(image string) velez_api.ServicePattern {
	return patternRepositories()[imageRepositoryName(image)]
}
