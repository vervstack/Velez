package domain

import (
	"errors"
	"slices"
	"strings"
)

// ErrInvalidSecretRef reports a string that does not parse into a SecretRef -
// see ParseSecretRef.
var ErrInvalidSecretRef = errors.New("invalid secret ref")

const (
	// secretRefSegments is the number of "/"-separated segments in a
	// canonical SecretRef string: scope/owner/key.
	secretRefSegments = 3
)

// SecretRef identifies one value in internal/service/secrets.Store - never the
// value itself. Other tables store its canonical string form, never a value:
// velez.pg_instances.secret_ref, velez.registries.secret. See
// docs/features/pgaas_and_registry_plugin.md section 1.
type SecretRef struct {
	Scope string
	Owner string
	Key   string
}

// String returns the canonical "scope/owner/key" form.
func (r SecretRef) String() string {
	return r.Scope + "/" + r.Owner + "/" + r.Key
}

// ParseSecretRef parses the canonical "scope/owner/key" form - the inverse of
// SecretRef.String. All three segments must be non-empty and the string must
// carry exactly two "/" separators, otherwise ErrInvalidSecretRef.
func ParseSecretRef(ref string) (SecretRef, error) {
	parts := strings.Split(ref, "/")
	if len(parts) != secretRefSegments {
		return SecretRef{}, ErrInvalidSecretRef
	}

	if slices.Contains(parts, "") {
		return SecretRef{}, ErrInvalidSecretRef
	}

	parsed := SecretRef{
		Scope: parts[0],
		Owner: parts[1],
		Key:   parts[2],
	}

	return parsed, nil
}
