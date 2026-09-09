package pgaas

import (
	"strings"
)

const (
	// pgPasswordLength is the byte length passed to toolbox.RandomBase64
	// when generating a fresh PG instance password - this package's own
	// choice, independent of internal/jobs/enable_statefull.go's
	// generatedPwdLength constant (which mints the cluster's root/node
	// passwords, a different credential entirely).
	pgPasswordLength = 20
)

// sanitizeIdentifier derives a valid Postgres identifier from a Velez
// service name (which may contain hyphens, uppercase letters, etc. that
// Postgres identifiers can't): lowercased, every non [a-z0-9_] byte replaced
// with '_', and prefixed if the result would start with a digit or be empty.
// Not truncated to Postgres's 63-byte identifier limit - Velez service names
// are validated short elsewhere and this is not a bound this package
// enforces.
func sanitizeIdentifier(name string) string {
	var b strings.Builder

	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}

	result := b.String()
	if result == "" || (result[0] >= '0' && result[0] <= '9') {
		result = "pg_" + result
	}

	return result
}

// pgVolumeName derives a per-instance Docker volume name from the instance's
// service name, so multiple PG instances launched from the same builtin
// descriptor (whose volume name is fixed to "postgres-data") don't collide
// in Docker's global volume namespace.
func pgVolumeName(instanceName string) string {
	return instanceName + "-data"
}
