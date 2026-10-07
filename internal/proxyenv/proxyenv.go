// Package proxyenv builds and parses the environment variables that point a
// container's outbound traffic at a proxy.
package proxyenv

import (
	"net/url"
	"slices"
	"strings"
)

const (
	bypassSeparator = ","

	httpsProxyUpper = "HTTPS_PROXY"
	httpsProxyLower = "https_proxy"
	noProxyUpper    = "NO_PROXY"
	noProxyLower    = "no_proxy"
)

var (
	// keys carries both spellings: curl and git read only the lowercase
	// http_proxy, most Go and Node clients read the uppercase one.
	keys = []string{
		"HTTP_PROXY", httpsProxyUpper, "ALL_PROXY",
		"http_proxy", httpsProxyLower, "all_proxy",
		noProxyUpper, noProxyLower,
	}

	proxyUrlKeys = []string{httpsProxyUpper, httpsProxyLower}
	bypassKeys   = []string{noProxyUpper, noProxyLower}

	defaultBypassHosts = []string{
		"localhost", "127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
	}

	supportedSchemes = []string{"http", "https", "socks5", "socks5h"}
)

// Keys returns every env key Env sets and Overrides clears.
func Keys() []string {
	return slices.Clone(keys)
}

// IsValidUrl reports whether raw is a proxy address with a supported scheme and a host.
func IsValidUrl(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}

	return slices.Contains(supportedSchemes, parsed.Scheme) && parsed.Host != ""
}

// Env returns the proxy env for proxyUrl, or an empty map when proxyUrl is empty.
// NO_PROXY is the default bypass hosts followed by bypassHosts, deduplicated in order.
func Env(proxyUrl string, bypassHosts []string) map[string]string {
	env := make(map[string]string, len(keys))

	if proxyUrl == "" {
		return env
	}

	noProxy := strings.Join(mergeBypassHosts(bypassHosts), bypassSeparator)

	for _, key := range keys {
		if slices.Contains(bypassKeys, key) {
			env[key] = noProxy

			continue
		}

		env[key] = proxyUrl
	}

	return env
}

// Overrides is Env, except an empty proxyUrl yields every key with an empty
// value - the "delete this key" convention of domain.UpgradeDeployReq.EnvOverrides.
func Overrides(proxyUrl string, bypassHosts []string) map[string]string {
	if proxyUrl != "" {
		return Env(proxyUrl, bypassHosts)
	}

	cleared := make(map[string]string, len(keys))

	for _, key := range keys {
		cleared[key] = ""
	}

	return cleared
}

// Parse reads the proxy back out of a container env: the proxy url and the
// bypass hosts that are not Velez's defaults.
func Parse(env map[string]string) (proxyUrl string, bypassHosts []string) {
	proxyUrl = firstValue(env, proxyUrlKeys)

	for host := range strings.SplitSeq(firstValue(env, bypassKeys), bypassSeparator) {
		host = strings.TrimSpace(host)

		if host == "" || slices.Contains(defaultBypassHosts, host) || slices.Contains(bypassHosts, host) {
			continue
		}

		bypassHosts = append(bypassHosts, host)
	}

	return proxyUrl, bypassHosts
}

// ParseList is Parse over a Docker-style `KEY=value` env list.
func ParseList(env []string) (proxyUrl string, bypassHosts []string) {
	byName := make(map[string]string, len(env))

	for _, entry := range env {
		name, value, isPair := strings.Cut(entry, "=")
		if isPair {
			byName[name] = value
		}
	}

	return Parse(byName)
}

func mergeBypassHosts(extra []string) []string {
	merged := slices.Clone(defaultBypassHosts)

	for _, host := range extra {
		host = strings.TrimSpace(host)

		if host == "" || slices.Contains(merged, host) {
			continue
		}

		merged = append(merged, host)
	}

	return merged
}

func firstValue(env map[string]string, candidates []string) string {
	for _, key := range candidates {
		if env[key] != "" {
			return env[key]
		}
	}

	return ""
}
