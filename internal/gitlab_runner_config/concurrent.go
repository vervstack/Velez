// Package gitlab_runner_config reads and rewrites the global keys of a
// gitlab-runner config.toml. Shared by the GitLab runner provider (writes) and
// single-node runner storage (reads), neither of which may import the other.
package gitlab_runner_config

import (
	"bytes"
	"regexp"
	"strconv"
)

const (
	// DataPath is gitlab-runner's config/registration directory; it must match
	// builtin/gitlab_runner/deployment.yaml's volume mount.
	DataPath = "/etc/gitlab-runner"

	ConfigPath = DataPath + "/config.toml"
)

var (
	tableHeaderLine    = regexp.MustCompile(`^\s*\[`)
	concurrentKeyLine  = regexp.MustCompile(`^\s*concurrent\s*=`)
	concurrentValue    = regexp.MustCompile(`^\s*concurrent\s*=\s*(\d+)`)
	lineTerminatorByte = []byte("\n")
)

// Concurrent returns config.toml's top-level `concurrent` key. The bool is
// false when the key is absent or not a number.
func Concurrent(config []byte) (int32, bool) {
	for _, line := range bytes.SplitAfter(config, lineTerminatorByte) {
		if tableHeaderLine.Match(line) {
			return 0, false
		}

		match := concurrentValue.FindSubmatch(line)
		if match == nil {
			continue
		}

		value, err := strconv.ParseInt(string(match[1]), 10, 32)
		if err != nil {
			return 0, false
		}

		return int32(value), true
	}

	return 0, false
}

// SetConcurrent sets config.toml's top-level `concurrent` key. Top-level keys
// precede the first [table]/[[table]] header, so a `concurrent` line inside a
// [[runners]] table is never touched. Everything else is preserved byte for
// byte.
func SetConcurrent(config []byte, concurrent int32) []byte {
	line := []byte("concurrent = " + strconv.FormatInt(int64(concurrent), 10))

	lines := bytes.SplitAfter(config, lineTerminatorByte)
	out := make([]byte, 0, len(config)+len(line)+1)

	isReplaced := false

	for i, current := range lines {
		if tableHeaderLine.Match(current) {
			out = insertIfMissing(out, line, isReplaced)
			out = append(out, bytes.Join(lines[i:], nil)...)

			return out
		}

		if !isReplaced && concurrentKeyLine.Match(current) {
			out = append(out, line...)
			out = append(out, lineEnding(current)...)
			isReplaced = true

			continue
		}

		out = append(out, current...)
	}

	return insertIfMissing(out, line, isReplaced)
}

// insertIfMissing puts the concurrent line first when the top-level section
// has none - only called once that section has been fully scanned.
func insertIfMissing(topLevel, line []byte, isReplaced bool) []byte {
	if isReplaced {
		return topLevel
	}

	inserted := make([]byte, 0, len(topLevel)+len(line)+1)

	inserted = append(inserted, line...)
	inserted = append(inserted, lineTerminatorByte...)

	return append(inserted, topLevel...)
}

func lineEnding(line []byte) []byte {
	if bytes.HasSuffix(line, []byte("\r\n")) {
		return []byte("\r\n")
	}

	if bytes.HasSuffix(line, lineTerminatorByte) {
		return lineTerminatorByte
	}

	return nil
}
