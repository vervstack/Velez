package container_manager

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func Test_SuggestedRunnerDefaults_Scenarios(t *testing.T) {
	cases := []struct {
		name  string
		image string
		env   map[string]string
		want  *velez_api.SuggestedRunnerDefaults
	}{
		{
			name:  "github runner image with repo env",
			image: "ghcr.io/actions/actions-runner:latest",
			env:   map[string]string{"RUNNER_REPO_URL": "https://github.com/acme/app", "RUNNER_TOKEN": "t"},
			want: &velez_api.SuggestedRunnerDefaults{
				Provider:                 "GITHUB",
				Scope:                    "REPO",
				Target:                   "acme/app",
				IsRegistrationTokenFound: true,
			},
		},
		{
			name:  "gitlab runner image with server url",
			image: "gitlab/gitlab-runner:latest",
			env:   map[string]string{"CI_SERVER_URL": "https://git.example.com"},
			want: &velez_api.SuggestedRunnerDefaults{
				Provider: "GITLAB",
				Scope:    "RUNNER_SCOPE_UNSPECIFIED",
				BaseUrl:  "https://git.example.com",
			},
		},
		{
			name:  "non runner image has no defaults",
			image: "postgres:16",
			env:   map[string]string{"RUNNER_TOKEN": "t"},
			want:  nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := suggestedRunnerDefaults(tc.image, nil, tc.env)

			require.Equal(t, tc.want.GetProvider(), got.GetProvider())
			require.Equal(t, tc.want.GetScope(), got.GetScope())
			require.Equal(t, tc.want.GetTarget(), got.GetTarget())
			require.Equal(t, tc.want.GetBaseUrl(), got.GetBaseUrl())
			require.Equal(t, tc.want.GetIsRegistrationTokenFound(), got.GetIsRegistrationTokenFound())
		})
	}
}
