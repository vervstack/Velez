package container_derived

import (
	"testing"

	"github.com/stretchr/testify/require"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

func Test_RunnerFactsFromContainer_Scenarios(t *testing.T) {
	cases := []struct {
		name     string
		provider pb.RunnerProvider
		labels   map[string]string
		env      map[string]string
		want     RunnerFacts
	}{
		{
			name:     "github repo url and token",
			provider: pb.RunnerProvider_GITHUB,
			env: map[string]string{
				"RUNNER_REPO_URL": "https://github.com/acme/app",
				githubEnvToken:    "tok",
				"RUNNER_LABELS":   "a,b",
			},
			want: RunnerFacts{
				Provider:          pb.RunnerProvider_GITHUB,
				Scope:             pb.RunnerScope_REPO,
				Target:            "acme/app",
				Labels:            []string{"a", "b"},
				RegistrationToken: "tok",
			},
		},
		{
			name:     "github org url",
			provider: pb.RunnerProvider_GITHUB,
			env:      map[string]string{"RUNNER_ORG_URL": "https://github.com/acme"},
			want: RunnerFacts{
				Provider: pb.RunnerProvider_GITHUB,
				Scope:    pb.RunnerScope_ORG,
				Target:   "acme",
			},
		},
		{
			name:     "gitlab server url tags and token",
			provider: pb.RunnerProvider_GITLAB,
			env: map[string]string{
				"CI_SERVER_URL":      "https://git.example.com",
				"REGISTRATION_TOKEN": "glrt",
				"RUNNER_TAG_LIST":    "docker,arm",
			},
			want: RunnerFacts{
				Provider:          pb.RunnerProvider_GITLAB,
				BaseUrl:           "https://git.example.com",
				Labels:            []string{"docker", "arm"},
				RegistrationToken: "glrt",
			},
		},
		{
			name:     "velez stamp wins over provider token",
			provider: pb.RunnerProvider_GITHUB,
			env:      map[string]string{"VELEZ_RUNNER_REGISTRATION_TOKEN": "stamped", githubEnvToken: "raw"},
			want:     RunnerFacts{Provider: pb.RunnerProvider_GITHUB, RegistrationToken: "stamped"},
		},
		{
			name:     "velez labels win over env and provider hint",
			provider: pb.RunnerProvider_GITLAB,
			labels: map[string]string{
				labels.RunnerProviderLabel: "GITHUB",
				labels.RunnerScopeLabel:    "ORG",
				labels.RunnerTargetLabel:   "labelled",
				labels.RunnerLabelsLabel:   "x",
			},
			env: map[string]string{"RUNNER_REPO_URL": "https://github.com/acme/app"},
			want: RunnerFacts{
				Provider: pb.RunnerProvider_GITHUB,
				Scope:    pb.RunnerScope_ORG,
				Target:   "labelled",
				Labels:   []string{"x"},
			},
		},
		{
			name:     "unspecified provider reads only the velez stamp",
			provider: pb.RunnerProvider_RUNNER_PROVIDER_UNSPECIFIED,
			env:      map[string]string{githubEnvToken: "raw"},
			want:     RunnerFacts{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RunnerFactsFromContainer(tc.provider, tc.labels, tc.env)

			require.Equal(t, tc.want, got)
		})
	}
}

func Test_RunnerRegistrationToken_Scenarios(t *testing.T) {
	cases := []struct {
		name     string
		provider pb.RunnerProvider
		env      map[string]string
		want     string
	}{
		{"github runner token", pb.RunnerProvider_GITHUB, map[string]string{githubEnvToken: "a"}, "a"},
		{"gitlab registration token", pb.RunnerProvider_GITLAB, map[string]string{"REGISTRATION_TOKEN": "b"}, "b"},
		{"no token", pb.RunnerProvider_GITHUB, map[string]string{"FOO": "c"}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, RunnerRegistrationToken(tc.provider, tc.env))
		})
	}
}
