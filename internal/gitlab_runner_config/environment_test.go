package gitlab_runner_config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testManagedKeys = []string{testHttpProxyKey, "http_proxy", "NO_PROXY"}
)

const (
	testHttpProxyKey = "HTTP_PROXY"
	testConcurrent   = "concurrent = 1\n"

	runnerWithDocker = "[[runners]]\n  name = \"x\"\n  executor = \"docker\"\n" +
		"  [runners.docker]\n    image = \"alpine\"\n"
)

func Test_SetRunnerEnvironment_Scenarios(t *testing.T) {
	entries := map[string]string{
		testHttpProxyKey: "socks5://p:1080",
		"http_proxy":     "socks5://p:1080",
		"NO_PROXY":       "localhost,docker",
	}

	const proxyLine = `environment = ["HTTP_PROXY=socks5://p:1080", "http_proxy=socks5://p:1080", ` +
		`"NO_PROXY=localhost,docker"]`

	cases := []struct {
		name    string
		config  string
		entries map[string]string
		want    string
	}{
		{
			name:    "inserts the key right under the runners header",
			config:  "concurrent = 1\n\n" + runnerWithDocker,
			entries: entries,
			want: "concurrent = 1\n\n[[runners]]\n  " + proxyLine + "\n  name = \"x\"\n  executor = \"docker\"\n" +
				"  [runners.docker]\n    image = \"alpine\"\n",
		},
		{
			name: "keeps foreign entries and replaces managed ones in place",
			config: "[[runners]]\n  name = \"x\"\n  environment = [\"FOO=bar\", \"HTTP_PROXY=old\", \"NO_PROXY=old\"]\n" +
				"  [runners.docker]\n    image = \"alpine\"\n",
			entries: entries,
			want: "[[runners]]\n  name = \"x\"\n  environment = [\"FOO=bar\", \"HTTP_PROXY=socks5://p:1080\", " +
				"\"http_proxy=socks5://p:1080\", \"NO_PROXY=localhost,docker\"]\n" +
				"  [runners.docker]\n    image = \"alpine\"\n",
		},
		{
			name: "reads a multi-line array",
			config: "[[runners]]\n  environment = [\n    \"FOO=bar\",\n    \"HTTP_PROXY=old\",\n  ]\n" +
				"  [runners.docker]\n    image = \"alpine\"\n",
			entries: entries,
			want: "[[runners]]\n  environment = [\"FOO=bar\", \"HTTP_PROXY=socks5://p:1080\", " +
				"\"http_proxy=socks5://p:1080\", \"NO_PROXY=localhost,docker\"]\n" +
				"  [runners.docker]\n    image = \"alpine\"\n",
		},
		{
			name: "removes managed entries and keeps the foreign ones",
			config: "[[runners]]\n  environment = [\"FOO=bar\", \"HTTP_PROXY=old\"]\n" +
				"  [runners.docker]\n    image = \"alpine\"\n",
			entries: nil,
			want:    "[[runners]]\n  environment = [\"FOO=bar\"]\n  [runners.docker]\n    image = \"alpine\"\n",
		},
		{
			name:    "deletes the key when nothing is left",
			config:  "[[runners]]\n  environment = [\"HTTP_PROXY=old\"]\n  name = \"x\"\n",
			entries: map[string]string{testHttpProxyKey: ""},
			want:    "[[runners]]\n  name = \"x\"\n",
		},
		{
			name:    "clearing a config without the key changes nothing",
			config:  "concurrent = 1\n\n" + runnerWithDocker,
			entries: nil,
			want:    "concurrent = 1\n\n" + runnerWithDocker,
		},
		{
			name:    "does not touch an environment key in a docker sub-table",
			config:  "[[runners]]\n  name = \"x\"\n  [runners.docker]\n    environment = \"nope\"\n",
			entries: map[string]string{testHttpProxyKey: "http://p:1"},
			want: "[[runners]]\n  environment = [\"HTTP_PROXY=http://p:1\"]\n  name = \"x\"\n" +
				"  [runners.docker]\n    environment = \"nope\"\n",
		},
		{
			name:    "same entries is a byte for byte no-op",
			config:  "[[runners]]\n  environment = [\"HTTP_PROXY=a\"]   # hand written\n",
			entries: map[string]string{testHttpProxyKey: "a"},
			want:    "[[runners]]\n  environment = [\"HTTP_PROXY=a\"]   # hand written\n",
		},
		{
			name:    "no runners entry and nothing to set is a no-op",
			config:  testConcurrent,
			entries: nil,
			want:    testConcurrent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SetRunnerEnvironment([]byte(tc.config), testManagedKeys, tc.entries)

			require.NoError(t, err)
			require.Equal(t, tc.want, string(got))
		})
	}
}

func Test_SetRunnerEnvironment_NoRunnersEntry_ReturnsError(t *testing.T) {
	_, err := SetRunnerEnvironment([]byte(testConcurrent), testManagedKeys, map[string]string{testHttpProxyKey: "a"})

	require.ErrorIs(t, err, ErrRunnerEntryMissing)
}

func Test_SetRunnerEnvironment_NotAnArray_ReturnsError(t *testing.T) {
	config := "[[runners]]\n  environment = \"FOO=bar\"\n"

	_, err := SetRunnerEnvironment([]byte(config), testManagedKeys, map[string]string{testHttpProxyKey: "a"})

	require.ErrorIs(t, err, ErrRunnerEnvironmentInvalid)
}
