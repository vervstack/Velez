// Package registryclients is a hand-rolled client for the Docker Registry
// HTTP API V2 (https://distribution.github.io/distribution/spec/api/),
// mirroring internal/clients/cluster_clients/headscale's shape - plain
// net/http, no library - but with HTTP basic auth instead of a bearer token.
package registryclients

import (
	"go.vervstack.ru/Velez/internal/domain"
)

// Client talks to one generic_v2 (Docker Registry HTTP API V2) registry.
type Client struct {
	baseURL  string
	username string
	secret   string
}

// New builds a Client for the given registry. Username may be empty - some
// private registries allow anonymous catalog reads.
func New(reg domain.Registry) *Client {
	return &Client{
		baseURL:  reg.Url,
		username: reg.Username,
		secret:   reg.Secret,
	}
}
