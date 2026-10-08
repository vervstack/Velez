package gitlab_runner_config

import (
	"slices"
	"strconv"

	"github.com/pelletier/go-toml/v2"
	"go.redsock.ru/rerrors"
)

const (
	runnerTokenKey = "token"
)

type runnerTokenDocument struct {
	Runners []runnerTokenEntry `toml:"runners"`
}

type runnerTokenEntry struct {
	Token *string `toml:"token"`
}

// RunnerToken returns the `token` key of the first `[[runners]]` entry. The
// bool is false when there is no entry, the key is absent or the file does not
// parse.
func RunnerToken(config []byte) (token string, isFound bool) {
	var document runnerTokenDocument

	err := toml.Unmarshal(config, &document)
	if err != nil {
		return "", false
	}

	if len(document.Runners) == 0 || document.Runners[0].Token == nil {
		return "", false
	}

	return *document.Runners[0].Token, true
}

// SetRunnerToken sets the `token` key of the first `[[runners]]` entry,
// inserting it right below the entry header when the entry has none. gitlab-runner
// rewrites this file when it rotates its token, so key order is not assumed;
// everything else is preserved byte for byte. Fails with ErrRunnerEntryMissing
// when there is no `[[runners]]` entry.
func SetRunnerToken(config []byte, token string) ([]byte, error) {
	lines := splitLines(config)

	headerIndex := slices.IndexFunc(lines, runnersHeaderLine.Match)
	if headerIndex < 0 {
		return nil, rerrors.Wrap(ErrRunnerEntryMissing)
	}

	endIndex := len(lines)

	for i := headerIndex + 1; i < len(lines); i++ {
		if tableHeaderLine.Match(lines[i]) {
			endIndex = i

			break
		}
	}

	body := lines[headerIndex+1 : endIndex]
	key := keyValue{key: runnerTokenKey, value: strconv.Quote(token), isSet: true}

	body, isFound := setKey(body, key)

	var missing [][]byte

	if !isFound {
		indent := dockerKeyIndent(lines[headerIndex], body)

		missing = append(missing, buildLine([]byte(indent), key, nil))
	}

	updated := make([][]byte, 0, len(lines)+len(missing))

	updated = append(updated, lines[:headerIndex+1]...)
	updated = append(updated, body...)
	updated = append(updated, lines[endIndex:]...)

	return joinLines(insertBlock(updated, headerIndex, missing)), nil
}
