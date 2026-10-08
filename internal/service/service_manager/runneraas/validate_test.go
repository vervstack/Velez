package runneraas

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testRunnerRepoTarget = "owner/repo"
	testRunnerOrgTarget  = "owner"
	testDockerSocketAddr = "tcp://host:2375"
)

func Test_ValidateRunnerTarget_Scenarios(t *testing.T) {
	cases := []struct {
		name    string
		scope   velez_api.RunnerScope
		target  string
		wantErr error
	}{
		{"valid repo", velez_api.RunnerScope_REPO, testRunnerRepoTarget, nil},
		{"valid org", velez_api.RunnerScope_ORG, testRunnerOrgTarget, nil},
		{"empty target", velez_api.RunnerScope_REPO, "", user_errors.ErrRunnerTargetEmpty},
		{
			"unspecified scope",
			velez_api.RunnerScope_RUNNER_SCOPE_UNSPECIFIED,
			testRunnerRepoTarget,
			user_errors.ErrRunnerScopeUnspecified,
		},
		{"repo missing slash", velez_api.RunnerScope_REPO, testRunnerOrgTarget, user_errors.ErrRunnerTargetInvalidRepoFormat},
		{"repo empty owner", velez_api.RunnerScope_REPO, "/repo", user_errors.ErrRunnerTargetInvalidRepoFormat},
		{"repo empty name", velez_api.RunnerScope_REPO, "owner/", user_errors.ErrRunnerTargetInvalidRepoFormat},
		{"repo nested path valid (gitlab group/subgroup/project)", velez_api.RunnerScope_REPO, "owner/repo/extra", nil},
		{"org contains slash", velez_api.RunnerScope_ORG, testRunnerRepoTarget, user_errors.ErrRunnerTargetInvalidOrgFormat},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRunnerTarget(tc.scope, tc.target)

			if tc.wantErr == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func Test_ValidateDockerSocketAddress_Scenarios(t *testing.T) {
	cases := []struct {
		name    string
		addr    string
		wantErr error
	}{
		{"empty falls back to host socket", "", nil},
		{"valid tcp address", testDockerSocketAddr, nil},
		{"unix socket rejected", "unix:///var/run/docker.sock", user_errors.ErrRunnerDockerSocketAddressInvalid},
		{"bare path rejected", "/var/run/docker.sock", user_errors.ErrRunnerDockerSocketAddressInvalid},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDockerSocketAddress(tc.addr)

			if tc.wantErr == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func Test_ValidateDockerSource_Scenarios(t *testing.T) {
	cases := []struct {
		name    string
		dind    string
		socket  string
		wantErr error
	}{
		{"dind only", "dind", "", nil},
		{"socket only", "", "tcp://host:2375", nil},
		{"neither", "", "", user_errors.ErrRunnerDockerSourceRequired},
		{"both", "dind", "tcp://host:2375", user_errors.ErrRunnerDockerSourceAmbiguous},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDockerSource(tc.dind, tc.socket)

			if tc.wantErr == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func Test_ValidateGitlabAccessToken_Scenarios(t *testing.T) {
	cases := []struct {
		name    string
		token   string
		wantErr error
	}{
		{"personal access token", "glpat-abcdef", nil},
		{"custom prefix personal access token", "corp-abcdef", nil},
		{"empty", "", user_errors.ErrGitlabAccessTokenEmpty},
		{"legacy registration token", "GR1348941abcdef", user_errors.ErrGitlabLegacyRegistrationToken},
		{"runner authentication token", "glrt-abcdef", user_errors.ErrGitlabLegacyRegistrationToken},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateGitlabAccessToken(tc.token)

			if tc.wantErr == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func Test_ValidateBuildkitRunner_Scenarios(t *testing.T) {
	cases := []struct {
		name      string
		provider  velez_api.RunnerProvider
		isDindSet bool
		wantErr   error
	}{
		{"a gitlab runner with a dind is allowed", velez_api.RunnerProvider_GITLAB, true, nil},
		{"a gitlab runner without a dind is rejected", velez_api.RunnerProvider_GITLAB, false,
			user_errors.ErrRunnerBuildkitRequiresDind},
		{"a github runner is rejected", velez_api.RunnerProvider_GITHUB, true,
			user_errors.ErrRunnerBuildkitRequiresDind},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateBuildkitRunner(tc.provider, tc.isDindSet)

			if tc.wantErr == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}
