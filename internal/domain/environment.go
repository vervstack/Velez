package domain

import (
	"time"
)

// Environment - an isolated deployment environment living on the same node(s).
//
// Name is what callers pass over the wire (CreateSmerd.Request.environment and
// friends); Suffix is what actually ends up in Docker naming/labels
// (labels.SuffixLabel), keeping containers/networks/volumes of different
// environments apart on a shared Docker host.
//
// NOTE: this is NOT ServiceEnvironment (internal/domain/graph.go) - that one is
// a per-service dashboard status projection and an unrelated concept.
type Environment struct {
	ID        int64
	Name      string
	Suffix    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateEnvironmentReq - payload for creating a new environment.
// An empty Suffix defaults to Name (see verv_services.CreateEnvironment).
type CreateEnvironmentReq struct {
	Name   string
	Suffix string
}

// UpdateEnvironmentReq - payload for updating an existing environment.
// Nil fields are left untouched.
type UpdateEnvironmentReq struct {
	ID     int64
	Name   *string
	Suffix *string
}

// DeleteEnvironmentReq - identifies the environment to delete either by ID or
// by Name (proto exposes both as optional).
type DeleteEnvironmentReq struct {
	ID   *int64
	Name *string
}
