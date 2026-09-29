package gitlab_runner_config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SetConcurrent_Scenarios(t *testing.T) {
	cases := []struct {
		name   string
		config string
		want   string
	}{
		{
			"inserts the key as the first line when absent",
			"log_level = \"info\"\n\n[[runners]]\n  name = \"x\"\n",
			"concurrent = 4\nlog_level = \"info\"\n\n[[runners]]\n  name = \"x\"\n",
		},
		{
			"replaces the key when present",
			"concurrent = 1\ncheck_interval = 0\n\n[[runners]]\n  name = \"x\"\n",
			"concurrent = 4\ncheck_interval = 0\n\n[[runners]]\n  name = \"x\"\n",
		},
		{
			"leaves a concurrent line inside a runners table untouched",
			"concurrent = 1\n\n[[runners]]\n  name = \"x\"\n  concurrent = 9\n",
			"concurrent = 4\n\n[[runners]]\n  name = \"x\"\n  concurrent = 9\n",
		},
		{
			"inserts at the top when only a runners table holds the key",
			"[[runners]]\n  concurrent = 9\n",
			"concurrent = 4\n[[runners]]\n  concurrent = 9\n",
		},
		{
			"replaces an indented key",
			"  concurrent   =   1\n",
			"concurrent = 4\n",
		},
		{
			"replaces a key on the last line without a trailing newline",
			"check_interval = 0\nconcurrent = 1",
			"check_interval = 0\nconcurrent = 4",
		},
		{"builds the key from empty input", "", "concurrent = 4\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SetConcurrent([]byte(tc.config), 4)

			require.Equal(t, tc.want, string(got))
		})
	}
}

func Test_Concurrent_Scenarios(t *testing.T) {
	cases := []struct {
		name   string
		config string
		want   int32
		isSet  bool
	}{
		{"reads the top-level key", "concurrent = 4\n\n[[runners]]\n  name = \"x\"\n", 4, true},
		{"reads an indented key on the last line", "check_interval = 0\n  concurrent   =   7", 7, true},
		{"ignores a key inside a runners table", "[[runners]]\n  concurrent = 9\n", 0, false},
		{"reports absent for empty input", "", 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, isSet := Concurrent([]byte(tc.config))

			require.Equal(t, tc.isSet, isSet)
			require.Equal(t, tc.want, got)
		})
	}
}
