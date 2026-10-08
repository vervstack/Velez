package gitlab_runner_config

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"go.redsock.ru/rerrors"
)

const (
	runnersTableName = "runners"
	dockerVolumesKey = "volumes"
	networkModeKey   = "network_mode"
)

var (
	// anyTableHeaderLine matches a `[table]` / `[[table]]` header line and
	// captures its name. Array-continuation lines never match: they carry
	// quotes, commas or `=`.
	anyTableHeaderLine = regexp.MustCompile(`^\s*\[\[?\s*([^\[\]=,"'#\s]+)\s*\]\]?\s*(?:#.*)?\s*$`)
)

// CountRunners returns the number of `[[runners]]` entries config.toml holds.
func CountRunners(config []byte) int {
	count := 0

	for _, line := range splitLines(config) {
		if runnersHeaderLine.Match(line) {
			count++
		}
	}

	return count
}

// RemoveRunners drops every `[[runners]]` entry together with its
// `[runners.*]` sub-tables. Everything outside them is preserved byte for byte.
func RemoveRunners(config []byte) []byte {
	lines := splitLines(config)
	kept := make([][]byte, 0, len(lines))

	isInsideRunners := false

	for _, line := range lines {
		match := anyTableHeaderLine.FindSubmatch(line)
		if match != nil {
			isInsideRunners = isRunnersTable(string(match[1]))
		}

		if !isInsideRunners {
			kept = append(kept, line)
		}
	}

	return joinLines(kept)
}

// DockerDefaults are the `[runners.docker]` keys a fresh registration is
// given; an empty slice leaves its key out.
type DockerDefaults struct {
	Volumes             []string
	PullPolicy          []string
	AllowedPullPolicies []string
}

// SetDockerDefaults sets d's keys in the `[runners.docker]` table, keeping the
// rest of the file untouched. Fails with ErrRunnerEntryMissing when there is
// no `[[runners]]` entry to hold the table.
func SetDockerDefaults(config []byte, d DockerDefaults) ([]byte, error) {
	keys := []keyValue{
		{key: dockerVolumesKey, value: quoteArray(d.Volumes), isSet: len(d.Volumes) > 0},
		{key: pullPolicyKey, value: quoteArray(d.PullPolicy), isSet: len(d.PullPolicy) > 0},
		{key: allowedPullPoliciesKey, value: quoteArray(d.AllowedPullPolicies), isSet: len(d.AllowedPullPolicies) > 0},
	}

	return applyDockerKeys(config, keys)
}

// SetNetworkMode sets `network_mode` in the `[runners.docker]` table - the
// docker network job containers join - keeping the rest of the file untouched.
// An empty networkMode deletes the key. Fails with ErrRunnerEntryMissing when
// the key has to be added and there is no `[[runners]]` entry to hold the table.
func SetNetworkMode(config []byte, networkMode string) ([]byte, error) {
	keys := []keyValue{
		{key: networkModeKey, value: strconv.Quote(networkMode), isSet: networkMode != ""},
	}

	return applyDockerKeys(config, keys)
}

type networkModeDocument struct {
	Runners []networkModeRunner `toml:"runners"`
}

type networkModeRunner struct {
	Docker networkModeDockerTable `toml:"docker"`
}

type networkModeDockerTable struct {
	NetworkMode string `toml:"network_mode"`
}

// NetworkMode returns `network_mode` of the first `[[runners]]` entry's
// `[runners.docker]` table; empty when the key or the entry is absent.
func NetworkMode(config []byte) (string, error) {
	var document networkModeDocument

	err := toml.Unmarshal(config, &document)
	if err != nil {
		return "", rerrors.Wrap(err, "error parsing gitlab-runner config.toml")
	}

	if len(document.Runners) == 0 {
		return "", nil
	}

	return document.Runners[0].Docker.NetworkMode, nil
}

func isRunnersTable(name string) bool {
	return name == runnersTableName || strings.HasPrefix(name, runnersTableName+".")
}
