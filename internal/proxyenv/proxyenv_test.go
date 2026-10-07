package proxyenv

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testProxyUrl     = "socks5://192.168.1.44:1080"
	testBypassHost   = "a.local"
	testBypassHostB  = "b.local"
	testDefaultBypss = "localhost,127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"
)

func Test_Env_Scenarios(t *testing.T) {
	cases := []struct {
		name        string
		proxyUrl    string
		bypassHosts []string
		wantNoProxy string
	}{
		{"defaults only", testProxyUrl, nil, testDefaultBypss},
		{"extra hosts are appended", testProxyUrl, []string{"gitlab.internal"}, testDefaultBypss + ",gitlab.internal"},
		{
			"duplicates and defaults are dropped, order kept",
			testProxyUrl,
			[]string{testBypassHostB, "localhost", testBypassHost, testBypassHostB, " ", ""},
			testDefaultBypss + "," + testBypassHostB + "," + testBypassHost,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := Env(tc.proxyUrl, tc.bypassHosts)

			require.Len(t, env, len(Keys()))

			proxyKeys := []string{"HTTP_PROXY", httpsProxyUpper, "ALL_PROXY", "http_proxy", httpsProxyLower, "all_proxy"}

			for _, key := range proxyKeys {
				require.Equal(t, tc.proxyUrl, env[key], key)
			}

			require.Equal(t, tc.wantNoProxy, env[noProxyUpper])
			require.Equal(t, tc.wantNoProxy, env[noProxyLower])
		})
	}
}

func Test_Env_EmptyProxyUrl_ReturnsNothing(t *testing.T) {
	require.Empty(t, Env("", []string{testBypassHost}))
}

func Test_Overrides_Scenarios(t *testing.T) {
	t.Run("a proxy url sets every key", func(t *testing.T) {
		require.Equal(t, Env(testProxyUrl, nil), Overrides(testProxyUrl, nil))
	})

	t.Run("an empty proxy url clears every key", func(t *testing.T) {
		overrides := Overrides("", []string{testBypassHost})

		require.Len(t, overrides, len(Keys()))

		for _, key := range Keys() {
			value, isPresent := overrides[key]

			require.True(t, isPresent, key)
			require.Empty(t, value, key)
		}
	})
}

func Test_Parse_Scenarios(t *testing.T) {
	cases := []struct {
		name         string
		env          map[string]string
		wantProxyUrl string
		wantBypass   []string
	}{
		{"no env", map[string]string{"PATH": "/bin"}, "", nil},
		{"round trip of defaults", Env(testProxyUrl, nil), testProxyUrl, nil},
		{
			"round trip of extra hosts",
			Env(testProxyUrl, []string{testBypassHost, testBypassHostB}),
			testProxyUrl,
			[]string{testBypassHost, testBypassHostB},
		},
		{
			"lowercase keys are enough",
			map[string]string{httpsProxyLower: testProxyUrl, noProxyLower: "localhost, a.local ,,a.local"},
			testProxyUrl,
			[]string{testBypassHost},
		},
		{
			"uppercase wins over lowercase",
			map[string]string{httpsProxyUpper: testProxyUrl, httpsProxyLower: "http://other:1"},
			testProxyUrl,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			proxyUrl, bypassHosts := Parse(tc.env)

			require.Equal(t, tc.wantProxyUrl, proxyUrl)
			require.Equal(t, tc.wantBypass, bypassHosts)
		})
	}
}

func Test_IsValidUrl_Scenarios(t *testing.T) {
	cases := []struct {
		raw  string
		want bool
	}{
		{"socks5://192.168.1.44:1080", true},
		{"socks5h://proxy:1080", true},
		{"http://user:pass@proxy:3128", true},
		{"https://proxy", true},
		{"", false},
		{"proxy:3128", false},
		{"ftp://proxy:21", false},
		{"socks5://", false},
		{"://bad", false},
	}

	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			require.Equal(t, tc.want, IsValidUrl(tc.raw))
		})
	}
}

func Test_ParseList_KeepsEqualsSignsInValues(t *testing.T) {
	env := []string{
		"PATH=/bin", httpsProxyUpper + "=http://user:pa=ss@proxy:3128", noProxyUpper + "=localhost,a.local", "BROKEN",
	}

	proxyUrl, bypassHosts := ParseList(env)

	require.Equal(t, "http://user:pa=ss@proxy:3128", proxyUrl)
	require.Equal(t, []string{testBypassHost}, bypassHosts)
}
