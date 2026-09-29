package gitlab

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
			got := setConcurrent([]byte(tc.config), 4)

			require.Equal(t, tc.want, string(got))
		})
	}
}
