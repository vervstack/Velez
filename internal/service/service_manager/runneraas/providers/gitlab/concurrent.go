package gitlab

import (
	"bytes"
	"regexp"
	"strconv"
)

var (
	tableHeaderLine    = regexp.MustCompile(`^\s*\[`)
	concurrentKeyLine  = regexp.MustCompile(`^\s*concurrent\s*=`)
	lineTerminatorByte = []byte("\n")
)

// setConcurrent sets config.toml's top-level `concurrent` key. Top-level keys
// precede the first [table]/[[table]] header, so a `concurrent` line inside a
// [[runners]] table is never touched. Everything else is preserved byte for
// byte.
func setConcurrent(config []byte, concurrent int32) []byte {
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
