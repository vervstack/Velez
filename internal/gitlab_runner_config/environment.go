package gitlab_runner_config

import (
	"bytes"
	"regexp"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"go.redsock.ru/rerrors"
)

const (
	environmentKey       = "environment"
	environmentSeparator = "="
)

var (
	environmentKeyLine = regexp.MustCompile(`^\s*environment\s*=(.*)$`)
)

type environmentDocument struct {
	Environment []string `toml:"environment"`
}

// SetRunnerEnvironment sets the managed keys in the first `[[runners]]` entry's
// runner-level `environment` array - the env gitlab-runner hands to every job
// container, not the one in `[runners.docker]`. Entries whose key is not in
// managedKeys are kept; entries gets written in managedKeys order, and a key
// with an empty or missing value is removed. The array is deleted when nothing
// is left in it. Fails with ErrRunnerEntryMissing when there is something to
// set and no `[[runners]]` entry to hold it.
func SetRunnerEnvironment(config []byte, managedKeys []string, entries map[string]string) ([]byte, error) {
	lines := splitLines(config)

	headerIndex := slices.IndexFunc(lines, runnersHeaderLine.Match)
	if headerIndex < 0 {
		return missingRunnerEnvironment(config, managedKeys, entries)
	}

	endIndex := len(lines)

	for i := headerIndex + 1; i < len(lines); i++ {
		if tableHeaderLine.Match(lines[i]) {
			endIndex = i

			break
		}
	}

	body := lines[headerIndex+1 : endIndex]

	existing, err := readEnvironmentEntries(body)
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading runner environment")
	}

	merged := mergeEnvironmentEntries(existing, managedKeys, entries)
	if slices.Equal(existing, merged) {
		return config, nil
	}

	key := keyValue{key: environmentKey, value: quoteArray(merged), isSet: len(merged) > 0}

	body, isFound := setKey(body, key)

	var missing [][]byte

	if !isFound && key.isSet {
		indent := dockerKeyIndent(lines[headerIndex], body)

		missing = append(missing, buildLine([]byte(indent), key, nil))
	}

	updated := make([][]byte, 0, len(lines)+len(missing))

	updated = append(updated, lines[:headerIndex+1]...)
	updated = append(updated, body...)
	updated = append(updated, lines[endIndex:]...)

	return joinLines(insertBlock(updated, headerIndex, missing)), nil
}

func missingRunnerEnvironment(config []byte, managedKeys []string, entries map[string]string) ([]byte, error) {
	if len(mergeEnvironmentEntries(nil, managedKeys, entries)) == 0 {
		return config, nil
	}

	return nil, rerrors.Wrap(ErrRunnerEntryMissing)
}

// readEnvironmentEntries parses the runner-level `environment` array out of
// the entry body; nil when the key is absent.
func readEnvironmentEntries(body [][]byte) ([]string, error) {
	for i, line := range body {
		match := environmentKeyLine.FindSubmatch(bytes.TrimRight(line, "\r\n"))
		if match == nil {
			continue
		}

		end := i
		if opensMultiLineArray(match[1]) {
			end = multiLineArrayEnd(body, i)
		}

		var document environmentDocument

		err := toml.Unmarshal(joinLines(body[i:end+1]), &document)
		if err != nil {
			return nil, rerrors.Wrap(ErrRunnerEnvironmentInvalid)
		}

		return document.Environment, nil
	}

	return nil, nil
}

func mergeEnvironmentEntries(existing, managedKeys []string, entries map[string]string) []string {
	merged := make([]string, 0, len(existing)+len(managedKeys))

	for _, entry := range existing {
		name, _, _ := strings.Cut(entry, environmentSeparator)

		if !slices.Contains(managedKeys, name) {
			merged = append(merged, entry)
		}
	}

	for _, name := range managedKeys {
		if entries[name] != "" {
			merged = append(merged, name+environmentSeparator+entries[name])
		}
	}

	if len(merged) == 0 {
		return nil
	}

	return merged
}
