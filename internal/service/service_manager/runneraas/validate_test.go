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
		{"repo extra slash", velez_api.RunnerScope_REPO, "owner/repo/extra", user_errors.ErrRunnerTargetInvalidRepoFormat},
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
		{"valid tcp address", "tcp://host:2375", nil},
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
