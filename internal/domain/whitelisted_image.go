package domain

import (
	"github.com/google/uuid"
)

// WhitelistedImage is a user-added image repository allowed to run with elevated access.
type WhitelistedImage struct {
	Id         uuid.UUID
	Repository string
}

// WhitelistedImageEntry is one row of the effective whitelist: a built-in repository (no Id) or a user-added one.
type WhitelistedImageEntry struct {
	Id         string
	Repository string
	IsBuiltin  bool
}
