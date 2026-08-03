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
	ID     int64
	Name   string
	Suffix string
	// DockerHost - endpoint of the dedicated Docker daemon serving this
	// environment. Empty (the only value produced today) means "the node's
	// shared daemon", i.e. isolation via Suffix labels/names alone.
	//
	// Nothing populates it yet: the velez.environments column backing it, and
	// the dedicatedRuntime that would consume it, both land with Phase 2 of
	// docs/container_runtimes/roadmap.md. It exists now so the runtime
	// resolver can branch on it and reject the not-yet-implemented tier
	// explicitly instead of silently serving it from the shared daemon.
	DockerHost string
	CreatedAt  time.Time
	UpdatedAt  time.Time
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
