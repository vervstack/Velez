package gitlab_runner_config

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
)

const (
	generatedConfig = `concurrent = 1
check_interval = 0
shutdown_timeout = 0

[session_server]
  session_timeout = 1800

[[runners]]
  name = "x"
  url = "https://gitlab.com"
  executor = "docker"
  [runners.docker]
    tls_verify = false
    image = "alpine:latest"
    privileged = true
`

	policyAlways       = "always"
	policyIfNotPresent = "if-not-present"
)

func fullSettings() domain.GitlabRunnerSettings {
	return domain.GitlabRunnerSettings{
		PullPolicy:          []string{policyAlways, policyIfNotPresent},
		AllowedPullPolicies: []string{policyAlways, policyIfNotPresent},
		CheckInterval:       3,
		LogLevel:            "debug",
		ShutdownTimeout:     30,
	}
}

func Test_ReadSettings_Cases(t *testing.T) {
	cases := []struct {
		name   string
		config string
		want   domain.GitlabRunnerSettings
	}{
		{
			name:   "string form pull policy",
			config: "[[runners]]\n  [runners.docker]\n    pull_policy = \"always\"\n",
			want:   domain.GitlabRunnerSettings{PullPolicy: []string{policyAlways}},
		},
		{
			name: "array form pull policy",
			config: "[[runners]]\n  [runners.docker]\n" +
				"    pull_policy = [\"always\", \"if-not-present\"]\n" +
				"    allowed_pull_policies = [\"always\"]\n",
			want: domain.GitlabRunnerSettings{
				PullPolicy:          []string{policyAlways, policyIfNotPresent},
				AllowedPullPolicies: []string{policyAlways},
			},
		},
		{
			name:   "absent keys",
			config: "concurrent = 1\n\n[[runners]]\n  name = \"x\"\n",
			want:   domain.GitlabRunnerSettings{},
		},
		{
			name:   "generated config keeps zero values",
			config: generatedConfig,
			want:   domain.GitlabRunnerSettings{},
		},
		{
			name: "all five keys",
			config: "check_interval = 3\nlog_level = \"debug\"\nshutdown_timeout = 30\n\n" +
				"[[runners]]\n  [runners.docker]\n" +
				"    pull_policy = [\"always\", \"if-not-present\"]\n" +
				"    allowed_pull_policies = [\"always\", \"if-not-present\"]\n",
			want: fullSettings(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ReadSettings([]byte(tc.config))
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_ReadSettings_InvalidToml(t *testing.T) {
	_, err := ReadSettings([]byte("check_interval = = 3"))
	require.Error(t, err)
}

func Test_ApplySettings_SetsAllKeysOnGeneratedConfig(t *testing.T) {
	got, err := ApplySettings([]byte(generatedConfig), fullSettings())
	require.NoError(t, err)

	want := `concurrent = 1
check_interval = 3
shutdown_timeout = 30
log_level = "debug"

[session_server]
  session_timeout = 1800

[[runners]]
  name = "x"
  url = "https://gitlab.com"
  executor = "docker"
  [runners.docker]
    pull_policy = ["always", "if-not-present"]
    allowed_pull_policies = ["always", "if-not-present"]
    tls_verify = false
    image = "alpine:latest"
    privileged = true
`
	require.Equal(t, want, string(got))
}

func Test_ApplySettings_ReplacesExistingKeysInPlace(t *testing.T) {
	config := "log_level = \"info\"\ncheck_interval = 9\n\n[[runners]]\n  [runners.docker]\n" +
		"    pull_policy = \"never\"\n    image = \"a\"\n"

	got, err := ApplySettings([]byte(config), fullSettings())
	require.NoError(t, err)

	want := "log_level = \"debug\"\ncheck_interval = 3\nshutdown_timeout = 30\n\n[[runners]]\n  [runners.docker]\n" +
		"    allowed_pull_policies = [\"always\", \"if-not-present\"]\n" +
		"    pull_policy = [\"always\", \"if-not-present\"]\n    image = \"a\"\n"
	require.Equal(t, want, string(got))
}

func Test_ApplySettings_ZeroSettingsDeleteKeys(t *testing.T) {
	configured, err := ApplySettings([]byte(generatedConfig), fullSettings())
	require.NoError(t, err)

	got, err := ApplySettings(configured, domain.GitlabRunnerSettings{})
	require.NoError(t, err)

	want := `concurrent = 1

[session_server]
  session_timeout = 1800

[[runners]]
  name = "x"
  url = "https://gitlab.com"
  executor = "docker"
  [runners.docker]
    tls_verify = false
    image = "alpine:latest"
    privileged = true
`
	require.Equal(t, want, string(got))
}

func Test_ApplySettings_IsIdempotent(t *testing.T) {
	configs := map[string]string{
		"generated":        generatedConfig,
		"missing docker":   "concurrent = 1\n\n[[runners]]\n  name = \"x\"",
		"multi-line array": "[[runners]]\n  [runners.docker]\n    pull_policy = [\n      \"always\",\n    ]\n",
	}

	for name, config := range configs {
		t.Run(name, func(t *testing.T) {
			once, err := ApplySettings([]byte(config), fullSettings())
			require.NoError(t, err)

			twice, err := ApplySettings(once, fullSettings())
			require.NoError(t, err)

			require.Equal(t, string(once), string(twice))
		})
	}
}

func Test_ApplySettings_CreatesDockerTableWhenMissing(t *testing.T) {
	config := "concurrent = 1\n\n[[runners]]\n  name = \"x\""

	got, err := ApplySettings([]byte(config), fullSettings())
	require.NoError(t, err)

	want := "concurrent = 1\ncheck_interval = 3\nlog_level = \"debug\"\nshutdown_timeout = 30\n\n" +
		"[[runners]]\n  name = \"x\"\n\n[runners.docker]\n" +
		"  pull_policy = [\"always\", \"if-not-present\"]\n" +
		"  allowed_pull_policies = [\"always\", \"if-not-present\"]\n"
	require.Equal(t, want, string(got))
}

func Test_ApplySettings_TopLevelOnlyConfigWithoutTables(t *testing.T) {
	s := domain.GitlabRunnerSettings{CheckInterval: 5}

	got, err := ApplySettings([]byte("concurrent = 1"), s)
	require.NoError(t, err)

	require.Equal(t, "concurrent = 1\ncheck_interval = 5\n", string(got))
}

func Test_ApplySettings_RunnerEntryMissing(t *testing.T) {
	s := domain.GitlabRunnerSettings{PullPolicy: []string{policyAlways}}

	_, err := ApplySettings([]byte("concurrent = 1\n"), s)
	require.ErrorIs(t, err, ErrRunnerEntryMissing)
}

func Test_ApplySettings_NoDockerKeysNeedsNoRunnerEntry(t *testing.T) {
	s := domain.GitlabRunnerSettings{LogLevel: "warn"}

	got, err := ApplySettings([]byte("concurrent = 1\n"), s)
	require.NoError(t, err)

	require.Equal(t, "concurrent = 1\nlog_level = \"warn\"\n", string(got))
}

func Test_ApplySettings_ReplacesMultiLineArray(t *testing.T) {
	config := "[[runners]]\n  [runners.docker]\n    pull_policy = [\n      \"never\",\n      \"always\"\n    ]\n" +
		"    image = \"a\"\n"
	s := domain.GitlabRunnerSettings{PullPolicy: []string{"if-not-present"}}

	got, err := ApplySettings([]byte(config), s)
	require.NoError(t, err)

	want := "[[runners]]\n  [runners.docker]\n    pull_policy = [\"if-not-present\"]\n    image = \"a\"\n"
	require.Equal(t, want, string(got))
}

func Test_ApplySettings_DeletesMultiLineArray(t *testing.T) {
	config := "[[runners]]\n  [runners.docker]\n    pull_policy = [\n      \"never\"\n    ]\n    image = \"a\"\n"

	got, err := ApplySettings([]byte(config), domain.GitlabRunnerSettings{})
	require.NoError(t, err)

	require.Equal(t, "[[runners]]\n  [runners.docker]\n    image = \"a\"\n", string(got))
}

func Test_ApplySettings_UnindentedDockerHeaderUsesTwoSpaces(t *testing.T) {
	config := "[[runners]]\n[runners.docker]\n"
	s := domain.GitlabRunnerSettings{PullPolicy: []string{policyAlways}}

	got, err := ApplySettings([]byte(config), s)
	require.NoError(t, err)

	require.Equal(t, "[[runners]]\n[runners.docker]\n  pull_policy = [\"always\"]\n", string(got))
}

func Test_ApplySettings_IgnoresKeysOutsideTheirSection(t *testing.T) {
	config := "[[runners]]\n  check_interval = 7\n  [runners.docker]\n    log_level = \"x\"\n"
	s := domain.GitlabRunnerSettings{CheckInterval: 2}

	got, err := ApplySettings([]byte(config), s)
	require.NoError(t, err)

	require.Equal(t, "check_interval = 2\n[[runners]]\n  check_interval = 7\n  [runners.docker]\n    log_level = \"x\"\n",
		string(got))
}

func Test_Settings_RoundTrip(t *testing.T) {
	cases := []struct {
		name string
		s    domain.GitlabRunnerSettings
	}{
		{"all set", fullSettings()},
		{"none set", domain.GitlabRunnerSettings{}},
		{"only docker", domain.GitlabRunnerSettings{PullPolicy: []string{"never"}}},
		{"only top level", domain.GitlabRunnerSettings{CheckInterval: 4, LogLevel: "warn"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			applied, err := ApplySettings([]byte(generatedConfig), tc.s)
			require.NoError(t, err)

			got, err := ReadSettings(applied)
			require.NoError(t, err)
			require.Equal(t, tc.s, got)
		})
	}
}
