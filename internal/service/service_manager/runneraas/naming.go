package runneraas

import (
	"strings"

	"github.com/google/uuid"
)

const (
	// runnerNameSuffixLen is the number of hex characters of a fresh
	// uuid.NewString() used to disambiguate the runner name - a repeated
	// CreateRunner call for the same target must not collide with a still-
	// running prior runner of the same name inside the provider's own
	// runner registry.
	runnerNameSuffixLen = 8
)

// runnerVolumeName derives a per-instance Docker volume name from the
// instance's service name, so multiple runners launched from the same
// builtin descriptor (whose volume name is fixed to a placeholder) don't
// collide in Docker's global volume namespace.
func runnerVolumeName(instanceName string) string {
	return instanceName + "-data"
}

// deriveRunnerName builds a unique-enough runner name from the target plus
// a short random suffix, so re-registering the same target never collides
// with a still-registered prior runner in the provider's own runner list.
func deriveRunnerName(target string) string {
	suffix := uuid.NewString()[:runnerNameSuffixLen]

	return sanitizeRunnerTarget(target) + "-" + suffix
}

// sanitizeRunnerTarget replaces every "/" with "-" and lowercases the
// result - the only character an owner/repo or owner target string can
// carry that a runner name cannot.
func sanitizeRunnerTarget(target string) string {
	return strings.ToLower(strings.ReplaceAll(target, "/", "-"))
}
