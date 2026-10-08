package jobs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func jobNames(namedJobs []NamedJob) []string {
	names := make([]string, 0, len(namedJobs))
	for _, namedJob := range namedJobs {
		names = append(names, namedJob.Name)
	}

	return names
}

func newSetRunnerBuildkitPayload(isEnabled bool) *velez_api.SetRunnerBuildkitTaskPayload {
	request := &velez_api.SetRunnerBuildkit_Request{Name: "gitlab-runner-x", IsBuildkitEnabled: isEnabled}

	return &velez_api.SetRunnerBuildkitTaskPayload{Request: request}
}

func Test_SetRunnerBuildkitHandler_BuildJobs_Scenarios(t *testing.T) {
	cases := []struct {
		name      string
		isEnabled bool
		want      []string
	}{
		{
			"enable installs binfmt, builds the network and buildkitd in the dind, points the runner at it, then marks it",
			true,
			[]string{
				stepInstallBinfmt, stepEnsureBuildkitNetwork, stepDeployBuildkit,
				stepWaitForBuildkit, stepSetBuildkitNetworkMode, stepMarkBuildkitEnabled,
			},
		},
		{
			"disable marks it off and points the runner away first, then drops buildkitd, network and volume",
			false,
			[]string{
				stepMarkBuildkitDisabled, stepClearBuildkitNetworkMode, stepRemoveBuildkit,
				stepRemoveBuildkitNetwork, stepRemoveBuildkitVolume,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewSetRunnerBuildkitHandler(&fakeClusterStorage{}, nil)

			namedJobs := h.BuildJobs(newSetRunnerBuildkitPayload(tc.isEnabled))

			require.Equal(t, tc.want, jobNames(namedJobs))
		})
	}
}

func Test_CreateRunnerHandler_BuildJobs_BuildkitSteps(t *testing.T) {
	cases := []struct {
		name      string
		isEnabled bool
		wantTail  []string
	}{
		{"without buildkit the chain ends at the runner row", false, []string{stepRegisterRunnerRow}},
		{"with buildkit the enable steps follow the runner row", true, []string{
			stepRegisterRunnerRow,
			stepInstallBinfmt, stepEnsureBuildkitNetwork, stepDeployBuildkit,
			stepWaitForBuildkit, stepSetBuildkitNetworkMode, stepMarkBuildkitEnabled,
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewCreateRunnerHandler(nil, &fakeClusterStorage{}, nil, nil, nil, nil)

			request := &velez_api.CreateRunner_Request{Name: "x", IsBuildkitEnabled: tc.isEnabled}
			payload := &velez_api.CreateRunnerTaskPayload{Request: request}

			names := jobNames(h.BuildJobs(payload))

			require.Equal(t, tc.wantTail, names[len(names)-len(tc.wantTail):])
		})
	}
}

func Test_BuildkitProxyEnv_Scenarios(t *testing.T) {
	t.Run("a runner without a proxy gives buildkitd none", func(t *testing.T) {
		require.Empty(t, buildkitProxyEnv([]string{"PATH=/bin"}))
	})

	t.Run("a runner proxy is copied to buildkitd", func(t *testing.T) {
		env := buildkitProxyEnv([]string{"PATH=/bin", "HTTPS_PROXY=socks5h://proxy:1080"})

		require.Contains(t, env, "HTTPS_PROXY=socks5h://proxy:1080")
		require.Contains(t, env, "https_proxy=socks5h://proxy:1080")
		require.NotContains(t, env, "PATH=/bin")
	})
}
