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
	Id     int64
	Name   string
	Suffix string
	// DockerHost - endpoint of the Docker daemon serving this environment.
	// Equal to the node's own configured Docker connection
	// (node_clients.Docker.Host()) means "the node's shared daemon" -
	// isolation via Suffix labels/names alone (container_runtime.resolver's
	// label-based tier). Any other value means a Docker daemon dedicated to
	// this environment (container_runtime.resolver's direct tier) - see
	// docs/container_runtimes/roadmap.md.
	//
	// NOT NULL: an environment with no explicit override defaults to the
	// node's own Docker host (verv_services.CreateEnvironment), never to an
	// empty string - there's no "unset" state, only "same as the node" or
	// "somewhere else".
	//
	// Only ever non-default under statefull mode - static (single-node/dev)
	// environments storage rejects every write
	// (user_errors.ErrRequiresStatefullMode), so this can only diverge from
	// the node's own host for an environment actually persisted in postgres.
	DockerHost string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// CreateEnvironmentReq - payload for creating a new environment.
// An empty Suffix defaults to Name, an empty DockerHost defaults to the
// node's own Docker host (see verv_services.CreateEnvironment).
type CreateEnvironmentReq struct {
	Name       string
	Suffix     string
	DockerHost string
}

// UpdateEnvironmentReq - payload for updating an existing environment.
// Nil fields are left untouched.
type UpdateEnvironmentReq struct {
	Id         int64
	Name       *string
	Suffix     *string
	DockerHost *string
}

// DeleteEnvironmentReq - identifies the environment to delete either by ID or
// by Name (proto exposes both as optional).
type DeleteEnvironmentReq struct {
	Id   *int64
	Name *string
}
