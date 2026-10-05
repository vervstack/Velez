package s3aas

import (
	"strconv"
	"strings"
)

const (
	regionKey            = "s3_region"
	replicationFactorKey = "replication_factor"
)

// parseGarageConfig reads the only two keys the instance view needs out of
// garage.toml; either is zero when absent.
func parseGarageConfig(content []byte) (region string, replicationFactor uint32) {
	for line := range strings.SplitSeq(string(content), "\n") {
		key, value, isAssignment := strings.Cut(stripComment(line), "=")
		if !isAssignment {
			continue
		}

		value = strings.TrimSpace(value)

		switch strings.TrimSpace(key) {
		case regionKey:
			region = strings.Trim(value, `"'`)
		case replicationFactorKey:
			parsed, err := strconv.ParseUint(value, 10, 32)
			if err == nil {
				replicationFactor = uint32(parsed)
			}
		}
	}

	return region, replicationFactor
}

func stripComment(line string) string {
	isQuoted := false

	for i, symbol := range line {
		switch symbol {
		case '"':
			isQuoted = !isQuoted
		case '#':
			if !isQuoted {
				return line[:i]
			}
		}
	}

	return line
}
