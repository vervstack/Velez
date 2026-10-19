package main

import (
	"context"
	"os"
	"strconv"
	"strings"

	"go.redsock.ru/rerrors"
)

const (
	envMaxGroups = "VELEZ_E2E_MAX_GROUPS"

	bytesPerGroup = 3 << 29
	minGroups     = 1
)

//nolint:forbidigo // package-private tool sentinel, not shared/user-facing
var errMaxGroupsInvalid = rerrors.New("max groups must be a positive integer")

// maxGroupsForMemory is how many groups fit in memTotalBytes of daemon memory.
func maxGroupsForMemory(memTotalBytes int64, groupCount int) int {
	fit := int(memTotalBytes / bytesPerGroup)

	return min(max(fit, minGroups), max(groupCount, minGroups))
}

func parseMaxGroups(raw string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value < 1 {
		return 0, rerrors.Wrap(errMaxGroupsInvalid, raw)
	}

	return value, nil
}

func dockerMemTotal(ctx context.Context) (int64, error) {
	out, err := docker(ctx, "info", "--format", "{{.MemTotal}}")
	if err != nil {
		return 0, rerrors.Wrap(err, "error reading docker memory: "+out)
	}

	memTotal, err := strconv.ParseInt(strings.TrimSpace(out), 10, 64)
	if err != nil {
		return 0, rerrors.Wrap(err, "error parsing docker memory: "+out)
	}

	return memTotal, nil
}

// resolveMaxGroups picks the concurrency cap: the flag, else the env var, else
// what the bootstrap docker daemon's memory allows.
func resolveMaxGroups(ctx context.Context, flagValue, groupCount int) (int, error) {
	if flagValue > 0 {
		return min(flagValue, groupCount), nil
	}

	raw := os.Getenv(envMaxGroups)
	if raw != "" {
		fromEnv, err := parseMaxGroups(raw)
		if err != nil {
			return 0, rerrors.Wrap(err, "error reading "+envMaxGroups)
		}

		return min(fromEnv, groupCount), nil
	}

	memTotal, err := dockerMemTotal(ctx)
	if err != nil {
		return 0, rerrors.Wrap(err, "error computing max groups")
	}

	return maxGroupsForMemory(memTotal, groupCount), nil
}
