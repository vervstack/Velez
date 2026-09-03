package domain

import (
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

// GetVervonomiconReq is the input to VervServicesService.GetVervonomicon.
type GetVervonomiconReq struct {
	ServiceName string

	// Environment - which <environment>/ overlay to merge, and which running
	// instance's image to resolve the descriptor from. Empty resolves the
	// default environment.
	Environment string
}

// VervonomiconResult is the output of VervServicesService.GetVervonomicon.
type VervonomiconResult struct {
	// NoDescriptor - true when the service has no .verv/ descriptor (or no
	// running instance to read one from) in this environment. Per
	// docs/features/vervonomicon.md, absence of a descriptor is never an
	// error - the transport layer renders a clean empty response, not a
	// gRPC error.
	NoDescriptor bool

	// Raw - every file found under .verv/, before overlay merging.
	Raw map[string][]byte

	// ResolvedYaml - the descriptor after the environment overlay is
	// merged in, re-marshalled to YAML, with config.sensitive values
	// masked.
	ResolvedYaml string

	Source      verv.SourceKind
	Environment string

	// ResourceStatuses - one entry per descriptor.Resources item, per
	// docs/features/vervonomicon.md's "Resource reconciliation": whether the
	// service already has a connection to that resource, or still needs one
	// provisioned. Empty when the descriptor declares no resources.
	ResourceStatuses []ResourceReconciliation
}

// ResourceConnectionStatus is one resources.yaml entry's reconciliation
// outcome - see docs/features/vervonomicon.md's "Resource reconciliation".
type ResourceConnectionStatus string

const (
	// ResourceAlreadyConnected - a velez.service_resources binding exists for
	// (service, name), or matreshka data_sources already holds a
	// non-placeholder connection for it. Velez creates nothing.
	ResourceAlreadyConnected ResourceConnectionStatus = "already_connected"

	// ResourceMustProvision - neither a binding nor a live matreshka
	// connection exists yet.
	ResourceMustProvision ResourceConnectionStatus = "must_provision"
)

// ResourceReconciliation is the reconciliation outcome for one resources.yaml
// entry.
type ResourceReconciliation struct {
	Name   string
	Type   string
	Status ResourceConnectionStatus
}

// CreateDeployFromVervonomiconReq is the input to
// VervServicesService.CreateDeployFromVervonomicon.
type CreateDeployFromVervonomiconReq struct {
	ServiceName string
	Environment string

	// Image - the image whose baked-in .verv/ descriptor drives this
	// deploy. Also wins over any app.image the descriptor itself sets, per
	// docs/features/vervonomicon.md's image-precedence rule.
	Image string
}
