package gitlab_runner_config

import (
	"bytes"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

const (
	checkIntervalKey       = "check_interval"
	logLevelKey            = "log_level"
	shutdownTimeoutKey     = "shutdown_timeout"
	pullPolicyKey          = "pull_policy"
	allowedPullPoliciesKey = "allowed_pull_policies"

	keyValueSeparator   = " = "
	defaultIndent       = "  "
	nestedIndent        = "    "
	defaultLineEnding   = "\n"
	dockerTableHeader   = "[runners.docker]"
	arrayClosingBracket = "]"
)

var (
	dockerHeaderLine  = regexp.MustCompile(`^\s*\[runners\.docker\]\s*$`)
	runnersHeaderLine = regexp.MustCompile(`^\s*\[\[runners\]\]\s*$`)
	leadingIndentLine = regexp.MustCompile(`^(\s*)[^\s#]`)
	headerIndentLine  = regexp.MustCompile(`^(\s*)\[`)
)

type settingsDocument struct {
	CheckInterval   int32         `toml:"check_interval"`
	LogLevel        string        `toml:"log_level"`
	ShutdownTimeout int32         `toml:"shutdown_timeout"`
	Runners         []runnerEntry `toml:"runners"`
}

type runnerEntry struct {
	Docker dockerTable `toml:"docker"`
}

type dockerTable struct {
	PullPolicy          []string `toml:"pull_policy"`
	AllowedPullPolicies []string `toml:"allowed_pull_policies"`
}

type singlePolicyDocument struct {
	CheckInterval   int32                `toml:"check_interval"`
	LogLevel        string               `toml:"log_level"`
	ShutdownTimeout int32                `toml:"shutdown_timeout"`
	Runners         []singlePolicyRunner `toml:"runners"`
}

type singlePolicyRunner struct {
	Docker singlePolicyDockerTable `toml:"docker"`
}

type singlePolicyDockerTable struct {
	PullPolicy          string   `toml:"pull_policy"`
	AllowedPullPolicies []string `toml:"allowed_pull_policies"`
}

// ReadSettings parses config.toml; a missing key yields the zero value.
func ReadSettings(config []byte) (domain.GitlabRunnerSettings, error) {
	var document settingsDocument

	err := toml.Unmarshal(config, &document)
	if err != nil {
		return readSinglePolicySettings(config, err)
	}

	settings := domain.GitlabRunnerSettings{
		CheckInterval:   document.CheckInterval,
		LogLevel:        document.LogLevel,
		ShutdownTimeout: document.ShutdownTimeout,
	}

	if len(document.Runners) > 0 {
		settings.PullPolicy = document.Runners[0].Docker.PullPolicy
		settings.AllowedPullPolicies = document.Runners[0].Docker.AllowedPullPolicies
	}

	return settings, nil
}

// readSinglePolicySettings retries the parse for a hand-written
// `pull_policy = "always"`, which TOML types as a string rather than an array.
func readSinglePolicySettings(config []byte, arrayErr error) (domain.GitlabRunnerSettings, error) {
	var document singlePolicyDocument

	err := toml.Unmarshal(config, &document)
	if err != nil {
		return domain.GitlabRunnerSettings{}, rerrors.Wrap(arrayErr, "error parsing gitlab-runner config.toml")
	}

	settings := domain.GitlabRunnerSettings{
		CheckInterval:   document.CheckInterval,
		LogLevel:        document.LogLevel,
		ShutdownTimeout: document.ShutdownTimeout,
	}

	if len(document.Runners) > 0 {
		docker := document.Runners[0].Docker

		if docker.PullPolicy != "" {
			settings.PullPolicy = []string{docker.PullPolicy}
		}

		settings.AllowedPullPolicies = docker.AllowedPullPolicies
	}

	return settings, nil
}

// ApplySettings sets all five keys to s; a zero value deletes its key.
// Everything else is preserved byte for byte.
func ApplySettings(config []byte, s domain.GitlabRunnerSettings) ([]byte, error) {
	withTopLevel := applyTopLevelSettings(config, s)

	withDocker, err := applyDockerSettings(withTopLevel, s)
	if err != nil {
		return nil, rerrors.Wrap(err, "error applying docker settings")
	}

	return withDocker, nil
}

func applyTopLevelSettings(config []byte, s domain.GitlabRunnerSettings) []byte {
	lines := splitLines(config)

	headerIndex := len(lines)

	for i, line := range lines {
		if tableHeaderLine.Match(line) {
			headerIndex = i

			break
		}
	}

	topLevel := lines[:headerIndex]

	var missing [][]byte

	keys := []keyValue{
		{key: checkIntervalKey, value: strconv.FormatInt(int64(s.CheckInterval), 10), isSet: s.CheckInterval != 0},
		{key: logLevelKey, value: strconv.Quote(s.LogLevel), isSet: s.LogLevel != ""},
		{key: shutdownTimeoutKey, value: strconv.FormatInt(int64(s.ShutdownTimeout), 10), isSet: s.ShutdownTimeout != 0},
	}

	for _, key := range keys {
		var isFound bool

		topLevel, isFound = setKey(topLevel, key)
		if !isFound && key.isSet {
			missing = append(missing, buildLine(nil, key, nil))
		}
	}

	merged := make([][]byte, 0, len(lines)+len(missing))

	merged = append(merged, topLevel...)
	merged = append(merged, lines[headerIndex:]...)

	anchor := lastContentLine(topLevel)

	return joinLines(insertBlock(merged, anchor, missing))
}

func applyDockerSettings(config []byte, s domain.GitlabRunnerSettings) ([]byte, error) {
	keys := []keyValue{
		{key: pullPolicyKey, value: quoteArray(s.PullPolicy), isSet: len(s.PullPolicy) > 0},
		{key: allowedPullPoliciesKey, value: quoteArray(s.AllowedPullPolicies), isSet: len(s.AllowedPullPolicies) > 0},
	}

	return applyDockerKeys(config, keys)
}

// applyDockerKeys sets (or, when unset, deletes) keys in the first
// `[runners.docker]` table, creating the table when it is missing.
func applyDockerKeys(config []byte, keys []keyValue) ([]byte, error) {
	lines := splitLines(config)

	headerIndex := -1

	for i, line := range lines {
		if dockerHeaderLine.Match(line) {
			headerIndex = i

			break
		}
	}

	if headerIndex < 0 {
		return appendDockerTable(config, lines, keys)
	}

	endIndex := len(lines)

	for i := headerIndex + 1; i < len(lines); i++ {
		if tableHeaderLine.Match(lines[i]) {
			endIndex = i

			break
		}
	}

	body := lines[headerIndex+1 : endIndex]
	indent := dockerKeyIndent(lines[headerIndex], body)

	var missing [][]byte

	for _, key := range keys {
		var isFound bool

		body, isFound = setKey(body, key)
		if !isFound && key.isSet {
			missing = append(missing, buildLine([]byte(indent), key, nil))
		}
	}

	merged := make([][]byte, 0, len(lines)+len(missing))

	merged = append(merged, lines[:headerIndex+1]...)
	merged = append(merged, body...)
	merged = append(merged, lines[endIndex:]...)

	return joinLines(insertBlock(merged, headerIndex, missing)), nil
}

func appendDockerTable(config []byte, lines [][]byte, keys []keyValue) ([]byte, error) {
	var keyLines [][]byte

	for _, key := range keys {
		if key.isSet {
			keyLines = append(keyLines, buildLine([]byte(defaultIndent), key, nil))
		}
	}

	if len(keyLines) == 0 {
		return config, nil
	}

	hasRunnersEntry := slices.ContainsFunc(lines, runnersHeaderLine.Match)
	if !hasRunnersEntry {
		return nil, rerrors.Wrap(ErrRunnerEntryMissing)
	}

	out := make([]byte, 0, len(config)+len(dockerTableHeader)+len(defaultLineEnding)*2)

	out = append(out, config...)

	if len(out) > 0 && !bytes.HasSuffix(out, lineTerminatorByte) {
		out = append(out, defaultLineEnding...)
	}

	out = append(out, defaultLineEnding...)
	out = append(out, dockerTableHeader...)
	out = append(out, defaultLineEnding...)

	for _, line := range keyLines {
		out = append(out, line...)
	}

	return out, nil
}

type keyValue struct {
	key   string
	value string
	isSet bool
}

// setKey replaces the first line carrying key in place (keeping its
// indentation and line ending), or deletes every such line when the key is
// unset. A multi-line array value counts as one line.
func setKey(lines [][]byte, kv keyValue) ([][]byte, bool) {
	pattern := regexp.MustCompile(`^(\s*)` + regexp.QuoteMeta(kv.key) + `\s*=(.*)$`)

	out := make([][]byte, 0, len(lines))
	isFound := false

	for i := 0; i < len(lines); i++ {
		match := pattern.FindSubmatch(bytes.TrimRight(lines[i], "\r\n"))
		if match == nil {
			out = append(out, lines[i])

			continue
		}

		end := i
		if opensMultiLineArray(match[2]) {
			end = multiLineArrayEnd(lines, i)
		}

		if kv.isSet && !isFound {
			out = append(out, buildLine(match[1], kv, lineEnding(lines[end])))
		}

		isFound = true
		i = end
	}

	return out, isFound
}

func opensMultiLineArray(value []byte) bool {
	trimmed := bytes.TrimSpace(value)

	return bytes.HasPrefix(trimmed, []byte("[")) && !bytes.Contains(trimmed, []byte(arrayClosingBracket))
}

func multiLineArrayEnd(lines [][]byte, start int) int {
	for i := start + 1; i < len(lines); i++ {
		if bytes.Contains(lines[i], []byte(arrayClosingBracket)) {
			return i
		}
	}

	return len(lines) - 1
}

func buildLine(indent []byte, kv keyValue, ending []byte) []byte {
	if ending == nil {
		ending = []byte(defaultLineEnding)
	}

	line := make([]byte, 0, len(indent)+len(kv.key)+len(kv.value)+len(ending)+len(keyValueSeparator))

	line = append(line, indent...)
	line = append(line, kv.key...)
	line = append(line, keyValueSeparator...)
	line = append(line, kv.value...)

	return append(line, ending...)
}

// insertBlock puts block after lines[afterIndex] (the very start when
// afterIndex is negative), terminating that anchor line first if it is the
// last line of a file without a trailing newline.
func insertBlock(lines [][]byte, afterIndex int, block [][]byte) [][]byte {
	if len(block) == 0 {
		return lines
	}

	out := make([][]byte, 0, len(lines)+len(block))

	out = append(out, lines[:afterIndex+1]...)

	if afterIndex >= 0 && lineEnding(lines[afterIndex]) == nil {
		terminated := append(append([]byte(nil), lines[afterIndex]...), defaultLineEnding...)

		out[afterIndex] = terminated
	}

	out = append(out, block...)

	return append(out, lines[afterIndex+1:]...)
}

func lastContentLine(lines [][]byte) int {
	for i, line := range slices.Backward(lines) {
		if len(bytes.TrimSpace(line)) > 0 {
			return i
		}
	}

	return -1
}

func dockerKeyIndent(header []byte, body [][]byte) string {
	for _, line := range body {
		match := leadingIndentLine.FindSubmatch(line)
		if match != nil {
			return string(match[1])
		}
	}

	headerMatch := headerIndentLine.FindSubmatch(header)
	if headerMatch != nil && len(headerMatch[1]) > 0 {
		return nestedIndent
	}

	return defaultIndent
}

func quoteArray(values []string) string {
	quoted := make([]string, 0, len(values))

	for _, value := range values {
		quoted = append(quoted, strconv.Quote(value))
	}

	return "[" + strings.Join(quoted, ", ") + "]"
}

func splitLines(config []byte) [][]byte {
	lines := bytes.SplitAfter(config, lineTerminatorByte)

	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	return lines
}

func joinLines(lines [][]byte) []byte {
	return bytes.Join(lines, nil)
}
