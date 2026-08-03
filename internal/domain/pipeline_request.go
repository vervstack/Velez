package domain

import (
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

type LaunchSmerd struct {
	*velez_api.CreateSmerd_Request

	// Suffix - the resolved environment suffix (labels.SuffixLabel value) for
	// CreateSmerd_Request.Environment. Threaded per call instead of being baked
	// into the pipeliner, so one Velez process can serve many environments.
	//
	// NOTE: this field is not round-tripped through a persisted deployment
	// specification - internal/workers/deploy_watcher.go unmarshals the stored
	// payload into the embedded *velez_api.CreateSmerd_Request only. The
	// watcher therefore re-resolves the suffix from the request's environment
	// name, which is the authoritative value on the wire.
	Suffix string `json:"-"`
}

type LaunchSmerdResult struct {
	ContainerId string
}
