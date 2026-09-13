package github

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testRunnerRepoTarget = "owner/repo"
	testRunnerOrgTarget  = "owner"
)

func Test_RegistrationTokenEndpoint_Scenarios(t *testing.T) {
	cases := []struct {
		name    string
		scope   velez_api.RunnerScope
		target  string
		want    string
		wantErr error
	}{
		{
			"repo scope",
			velez_api.RunnerScope_REPO,
			testRunnerRepoTarget,
			"https://api.github.com/repos/owner/repo/actions/runners/registration-token",
			nil,
		},
		{
			"org scope",
			velez_api.RunnerScope_ORG,
			testRunnerOrgTarget,
			"https://api.github.com/orgs/owner/actions/runners/registration-token",
			nil,
		},
		{
			"unspecified scope",
			velez_api.RunnerScope_RUNNER_SCOPE_UNSPECIFIED,
			testRunnerOrgTarget,
			"",
			user_errors.ErrRunnerScopeUnspecified,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := registrationTokenEndpoint(tc.scope, tc.target)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_ParseRegistrationTokenResponse_Scenarios(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		want    string
		wantErr error
	}{
		{"valid token", `{"token":"AABBCC","expires_at":"2026-01-01T00:00:00Z"}`, "AABBCC", nil},
		{"missing token field", `{"expires_at":"2026-01-01T00:00:00Z"}`, "", user_errors.ErrGithubRegistrationTokenMissing},
		{"empty token value", `{"token":""}`, "", user_errors.ErrGithubRegistrationTokenMissing},
		{"malformed json", `not json`, "", nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseRegistrationTokenResponse([]byte(tc.body))

			if tc.name == "malformed json" {
				require.Error(t, err)

				return
			}

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_RegistrationEnv_Scenarios(t *testing.T) {
	provider := New()

	cases := []struct {
		name   string
		scope  velez_api.RunnerScope
		target string
		want   map[string]string
	}{
		{
			"repo scope sets repo url",
			velez_api.RunnerScope_REPO,
			testRunnerRepoTarget,
			map[string]string{envRunnerRepoUrl: "https://github.com/owner/repo"},
		},
		{
			"org scope sets org url",
			velez_api.RunnerScope_ORG,
			testRunnerOrgTarget,
			map[string]string{envRunnerOrgUrl: "https://github.com/owner"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := provider.RegistrationEnv(tc.scope, tc.target, "runner-1", "token-1", []string{"self-hosted"})

			for key, want := range tc.want {
				require.Equal(t, want, env[key])
			}

			require.Equal(t, "token-1", env[envRunnerToken])
			require.Equal(t, "runner-1", env[envRunnerName])
			require.Equal(t, "self-hosted", env[envRunnerLabels])
		})
	}
}
