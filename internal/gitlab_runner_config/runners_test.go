package gitlab_runner_config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testNetworkMode = "r-buildkit"
	bareRunnerEntry = "[[runners]]\n  name = \"x\"\n"

	runnerEntryWithDocker = "[[runners]]\n  name = \"x\"\n  executor = \"docker\"\n" +
		"  [runners.docker]\n    image = \"alpine\"\n    volumes = [\"/cache\"]\n" +
		"  [runners.cache]\n    Type = \"s3\"\n"
)

func Test_CountRunners_Scenarios(t *testing.T) {
	cases := []struct {
		name   string
		config string
		want   int
	}{
		{"empty config", "", 0},
		{"global keys only", testConcurrent, 0},
		{"one entry", "concurrent = 1\n\n" + runnerEntryWithDocker, 1},
		{"five entries", "concurrent = 1\n" + runnerEntryWithDocker + runnerEntryWithDocker +
			runnerEntryWithDocker + runnerEntryWithDocker + runnerEntryWithDocker, 5},
		{"a runners sub-table is not an entry", "[runners.docker]\n  image = \"a\"\n", 0},
		{"header with trailing comment spacing", "  [[runners]]  \n", 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, CountRunners([]byte(tc.config)))
		})
	}
}

func Test_RemoveRunners_Scenarios(t *testing.T) {
	cases := []struct {
		name   string
		config string
		want   string
	}{
		{"empty config", "", ""},
		{
			"keeps a config without entries byte for byte",
			"concurrent = 1\ncheck_interval = 0\n",
			"concurrent = 1\ncheck_interval = 0\n",
		},
		{
			"drops an entry with all its sub-tables",
			"concurrent = 1\n\n" + runnerEntryWithDocker,
			"concurrent = 1\n\n",
		},
		{
			"drops every entry",
			"concurrent = 1\n" + runnerEntryWithDocker + runnerEntryWithDocker,
			testConcurrent,
		},
		{
			"keeps unrelated tables on both sides of an entry",
			"concurrent = 1\n[session_server]\n  session_timeout = 1800\n[[runners]]\n  name = \"x\"\n" +
				"[other]\n  key = 1\n",
			"concurrent = 1\n[session_server]\n  session_timeout = 1800\n[other]\n  key = 1\n",
		},
		{
			"keeps a multi-line array inside an entry from ending it early",
			"[[runners]]\n  [runners.docker]\n    volumes = [\n      \"/cache\",\n    ]\n  name = \"x\"\n",
			"",
		},
		{
			"handles a header with a trailing comment",
			"[[runners]] # first\n  name = \"x\"\n",
			"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RemoveRunners([]byte(tc.config))

			require.Equal(t, tc.want, string(got))
		})
	}
}

func Test_SetDockerDefaults_Scenarios(t *testing.T) {
	defaults := DockerDefaults{
		Volumes:             []string{"runner-cache:/cache"},
		PullPolicy:          []string{policyIfNotPresent},
		AllowedPullPolicies: []string{policyIfNotPresent, policyAlways},
	}

	cases := []struct {
		name   string
		config string
		want   string
	}{
		{
			"replaces the anonymous cache volume and adds the policies",
			"[[runners]]\n  [runners.docker]\n    image = \"alpine\"\n    volumes = [\"/cache\"]\n",
			"[[runners]]\n  [runners.docker]\n" +
				"    pull_policy = [\"if-not-present\"]\n" +
				"    allowed_pull_policies = [\"if-not-present\", \"always\"]\n" +
				"    image = \"alpine\"\n    volumes = [\"runner-cache:/cache\"]\n",
		},
		{
			"creates the docker table when the entry has none",
			"[[runners]]\n  name = \"x\"\n",
			"[[runners]]\n  name = \"x\"\n\n[runners.docker]\n" +
				"  volumes = [\"runner-cache:/cache\"]\n" +
				"  pull_policy = [\"if-not-present\"]\n" +
				"  allowed_pull_policies = [\"if-not-present\", \"always\"]\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SetDockerDefaults([]byte(tc.config), defaults)

			require.NoError(t, err)
			require.Equal(t, tc.want, string(got))
		})
	}
}

func Test_SetDockerDefaults_WithoutRunnerEntry_Fails(t *testing.T) {
	defaults := DockerDefaults{Volumes: []string{"runner-cache:/cache"}}

	_, err := SetDockerDefaults([]byte("concurrent = 1\n"), defaults)

	require.ErrorIs(t, err, ErrRunnerEntryMissing)
}

func Test_SetNetworkMode_Scenarios(t *testing.T) {
	cases := []struct {
		name        string
		config      string
		networkMode string
		want        string
	}{
		{
			"adds the key to an existing docker table",
			"[[runners]]\n  [runners.docker]\n    image = \"alpine\"\n",
			testNetworkMode,
			"[[runners]]\n  [runners.docker]\n    network_mode = \"r-buildkit\"\n    image = \"alpine\"\n",
		},
		{
			"creates the docker table when the entry has none",
			bareRunnerEntry,
			testNetworkMode,
			bareRunnerEntry + "\n[runners.docker]\n  network_mode = \"r-buildkit\"\n",
		},
		{
			"replaces an existing value in place",
			"[[runners]]\n  [runners.docker]\n    network_mode = \"old\"\n    image = \"alpine\"\n",
			testNetworkMode,
			"[[runners]]\n  [runners.docker]\n    network_mode = \"r-buildkit\"\n    image = \"alpine\"\n",
		},
		{
			"deletes the key when the value is empty",
			"[[runners]]\n  [runners.docker]\n    network_mode = \"old\"\n    image = \"alpine\"\n",
			"",
			"[[runners]]\n  [runners.docker]\n    image = \"alpine\"\n",
		},
		{
			"keeps everything else byte for byte",
			"concurrent = 1\n\n" + runnerEntryWithDocker,
			testNetworkMode,
			"concurrent = 1\n\n[[runners]]\n  name = \"x\"\n  executor = \"docker\"\n" +
				"  [runners.docker]\n    network_mode = \"r-buildkit\"\n    image = \"alpine\"\n" +
				"    volumes = [\"/cache\"]\n  [runners.cache]\n    Type = \"s3\"\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SetNetworkMode([]byte(tc.config), tc.networkMode)

			require.NoError(t, err)
			require.Equal(t, tc.want, string(got))
		})
	}
}

func Test_SetNetworkMode_WithoutRunnerEntry_Fails(t *testing.T) {
	_, err := SetNetworkMode([]byte("concurrent = 1\n"), testNetworkMode)

	require.ErrorIs(t, err, ErrRunnerEntryMissing)
}
