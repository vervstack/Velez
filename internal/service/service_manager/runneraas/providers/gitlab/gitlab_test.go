package gitlab

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testRunnerRepoTarget = "owner/repo"
)

func Test_MintRegistrationToken_Scenarios(t *testing.T) {
	cases := []struct {
		name        string
		accessToken string
		want        string
		wantErr     error
	}{
		{"non-empty token is returned unchanged", "glpat-AABBCC", "glpat-AABBCC", nil},
		{"empty token is rejected", "", "", user_errors.ErrGitlabAccessTokenEmpty},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provider := New()

			got, err := provider.MintRegistrationToken(
				context.Background(), velez_api.RunnerScope_REPO, testRunnerRepoTarget, "", tc.accessToken,
			)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_RegistrationEnv_ReturnsEmpty(t *testing.T) {
	provider := New()

	env := provider.RegistrationEnv(velez_api.RunnerScope_REPO, testRunnerRepoTarget, "", "runner-1", "token-1", nil)

	require.Empty(t, env)
}

func Test_RegisterCommand_Scenarios(t *testing.T) {
	cases := []struct {
		name    string
		baseUrl string
		want    string
	}{
		{
			"defaults to gitlab.com when base url is empty",
			"",
			"gitlab-runner register --non-interactive --url https://gitlab.com --registration-token " +
				"token-1 --executor docker --docker-image alpine:latest",
		},
		{
			"uses the given base url",
			"https://gitlab.example.com",
			"gitlab-runner register --non-interactive --url https://gitlab.example.com --registration-token " +
				"token-1 --executor docker --docker-image alpine:latest",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provider := New()

			got := provider.RegisterCommand(tc.baseUrl, "token-1")

			require.Equal(t, tc.want, got)
		})
	}
}
