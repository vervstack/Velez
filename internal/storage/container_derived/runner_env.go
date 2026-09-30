package container_derived

import (
	"strings"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	githubUrlPrefix = "https://github.com/"

	githubEnvRepoUrl = "RUNNER_REPO_URL"
	githubEnvOrgUrl  = "RUNNER_ORG_URL"
	githubEnvToken   = "RUNNER_TOKEN"
	githubEnvLabels  = "RUNNER_LABELS"

	gitlabEnvServerUrl = "CI_SERVER_URL"
	gitlabEnvToken     = "REGISTRATION_TOKEN"
	gitlabEnvTokenAlt  = "RUNNER_TOKEN"
	gitlabEnvTagList   = "RUNNER_TAG_LIST"
	gitlabEnvTagListCi = "CI_RUNNER_TAGS"
)

// RunnerFacts are the runner register values a container states in its
// labels and env. RegistrationToken is a secret: never put it in a response.
type RunnerFacts struct {
	Provider          pb.RunnerProvider
	Scope             pb.RunnerScope
	Target            string
	BaseUrl           string
	Labels            []string
	RegistrationToken string
}

// RunnerFactsFromContainer reads the facts out of the container's velez.runner*
// labels first, then the provider's own env keys for whatever they leave
// blank. provider is the caller's choice (e.g. suggested from the image) and
// is overridden by the provider label.
func RunnerFactsFromContainer(
	provider pb.RunnerProvider, containerLabels map[string]string, env map[string]string,
) RunnerFacts {
	facts := RunnerFacts{Provider: provider}

	labelProvider, isKnown := pb.RunnerProvider_value[containerLabels[labels.RunnerProviderLabel]]
	if isKnown && labelProvider != 0 {
		facts.Provider = pb.RunnerProvider(labelProvider)
	}

	labelScope := pb.RunnerScope_value[containerLabels[labels.RunnerScopeLabel]]

	facts.Scope = pb.RunnerScope(labelScope)
	facts.Target = containerLabels[labels.RunnerTargetLabel]
	facts.BaseUrl = containerLabels[labels.RunnerBaseUrlLabel]
	facts.Labels = splitRunnerLabels(containerLabels[labels.RunnerLabelsLabel])

	switch facts.Provider {
	case pb.RunnerProvider_GITHUB:
		fillGithubFacts(&facts, env)
	case pb.RunnerProvider_GITLAB:
		fillGitlabFacts(&facts, env)
	case pb.RunnerProvider_RUNNER_PROVIDER_UNSPECIFIED:
	}

	if facts.RegistrationToken == "" {
		facts.RegistrationToken = env[RunnerRegistrationTokenEnvVar]
	}

	return facts
}

// RunnerRegistrationToken is the registration token a container carries in its
// env: Velez's own stamp first, then the provider's key. Empty when none.
func RunnerRegistrationToken(provider pb.RunnerProvider, env map[string]string) string {
	facts := RunnerFactsFromContainer(provider, nil, env)

	return facts.RegistrationToken
}

func fillGithubFacts(facts *RunnerFacts, env map[string]string) {
	repoUrl := env[githubEnvRepoUrl]
	orgUrl := env[githubEnvOrgUrl]

	if facts.Target == "" && repoUrl != "" {
		facts.Target = strings.TrimPrefix(repoUrl, githubUrlPrefix)
		facts.Scope = pb.RunnerScope_REPO
	}

	if facts.Target == "" && orgUrl != "" {
		facts.Target = strings.TrimPrefix(orgUrl, githubUrlPrefix)
		facts.Scope = pb.RunnerScope_ORG
	}

	if len(facts.Labels) == 0 {
		facts.Labels = splitRunnerLabels(env[githubEnvLabels])
	}

	facts.RegistrationToken = firstNonEmpty(env[RunnerRegistrationTokenEnvVar], env[githubEnvToken])
}

func fillGitlabFacts(facts *RunnerFacts, env map[string]string) {
	if facts.BaseUrl == "" {
		facts.BaseUrl = env[gitlabEnvServerUrl]
	}

	if len(facts.Labels) == 0 {
		facts.Labels = splitRunnerLabels(firstNonEmpty(env[gitlabEnvTagList], env[gitlabEnvTagListCi]))
	}

	facts.RegistrationToken = firstNonEmpty(
		env[RunnerRegistrationTokenEnvVar], env[gitlabEnvToken], env[gitlabEnvTokenAlt])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
