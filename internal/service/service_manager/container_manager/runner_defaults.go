package container_manager

import (
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage/container_derived"
)

// suggestedRunnerDefaults is nil unless the image suggests a runner pattern.
func suggestedRunnerDefaults(
	image string, containerLabels, env map[string]string,
) *velez_api.SuggestedRunnerDefaults {
	var provider velez_api.RunnerProvider

	switch suggestedPattern(image) {
	case velez_api.ServicePattern_SERVICE_PATTERN_GITHUB_RUNNER:
		provider = velez_api.RunnerProvider_GITHUB
	case velez_api.ServicePattern_SERVICE_PATTERN_GITLAB_RUNNER:
		provider = velez_api.RunnerProvider_GITLAB
	case velez_api.ServicePattern_SERVICE_PATTERN_UNSPECIFIED,
		velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES,
		velez_api.ServicePattern_SERVICE_PATTERN_REGISTRY:
		return nil
	}

	facts := container_derived.RunnerFactsFromContainer(provider, containerLabels, env)

	return &velez_api.SuggestedRunnerDefaults{
		Provider:                 facts.Provider.String(),
		Scope:                    facts.Scope.String(),
		Target:                   facts.Target,
		BaseUrl:                  facts.BaseUrl,
		Labels:                   facts.Labels,
		IsRegistrationTokenFound: facts.RegistrationToken != "",
	}
}
