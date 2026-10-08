package gitlab_runner_config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	tokenConfig = "concurrent = 1\n\n[[runners]]\n  name = \"one\" # keep\n  token = \"glrt-old\"\n" +
		"  token_obtained_at = 2026-10-08T00:00:00Z\n  executor = \"docker\"\n" +
		"  [runners.docker]\n    image = \"alpine\"\n"

	twoRunnersConfig = "[[runners]]\n  name = \"one\"\n  token = \"glrt-one\"\n" +
		"[[runners]]\n  name = \"two\"\n  token = \"glrt-two\"\n"
)

func Test_RunnerToken_ReadsFirstEntry(t *testing.T) {
	token, isFound := RunnerToken([]byte(twoRunnersConfig))

	require.True(t, isFound)
	require.Equal(t, "glrt-one", token)
}

func Test_RunnerToken_MissingCases(t *testing.T) {
	cases := []struct {
		name   string
		config string
	}{
		{"no runners entry", "concurrent = 1\n"},
		{"entry without token key", "[[runners]]\n  name = \"one\"\n"},
		{"empty file", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, isFound := RunnerToken([]byte(tc.config))

			require.False(t, isFound)
		})
	}
}

func Test_SetRunnerToken_ReplacesOnlyTheToken(t *testing.T) {
	out, err := SetRunnerToken([]byte(tokenConfig), "glrt-new")

	require.NoError(t, err)
	require.Equal(t,
		"concurrent = 1\n\n[[runners]]\n  name = \"one\" # keep\n  token = \"glrt-new\"\n"+
			"  token_obtained_at = 2026-10-08T00:00:00Z\n  executor = \"docker\"\n"+
			"  [runners.docker]\n    image = \"alpine\"\n",
		string(out))
}

func Test_SetRunnerToken_InsertsWhenKeyAbsent(t *testing.T) {
	out, err := SetRunnerToken([]byte("[[runners]]\n  name = \"one\"\n"), "glrt-new")

	require.NoError(t, err)
	require.Equal(t, "[[runners]]\n  token = \"glrt-new\"\n  name = \"one\"\n", string(out))
}

func Test_SetRunnerToken_InsertKeepsIndentation(t *testing.T) {
	out, err := SetRunnerToken([]byte("  [[runners]]\n"), "glrt-new")

	require.NoError(t, err)
	require.Equal(t, "  [[runners]]\n    token = \"glrt-new\"\n", string(out))
}

func Test_SetRunnerToken_LeavesSecondEntryUntouched(t *testing.T) {
	out, err := SetRunnerToken([]byte(twoRunnersConfig), "glrt-new")

	require.NoError(t, err)
	require.Equal(t,
		"[[runners]]\n  name = \"one\"\n  token = \"glrt-new\"\n[[runners]]\n  name = \"two\"\n  token = \"glrt-two\"\n",
		string(out))

	token, isFound := RunnerToken(out)
	require.True(t, isFound)
	require.Equal(t, "glrt-new", token)
}

func Test_SetRunnerToken_NoRunnersEntry(t *testing.T) {
	_, err := SetRunnerToken([]byte("concurrent = 1\n"), "glrt-new")

	require.ErrorIs(t, err, ErrRunnerEntryMissing)
}
