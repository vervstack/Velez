package domain

import (
	"time"
)

// RegistryInstance is the registry-specific satellite row for a
// Container-Registry-as-a-Service instance (velez.registry_instances).
// Status, environment, image, container id and deploy history live on the
// underlying velez.services / deployment_specifications / deployments rows,
// read through VervServicesService and ListSmerds - never duplicated here.
// Mirrors PgInstance exactly. See
// docs/features/pgaas_and_registry_plugin.md section 3.
type RegistryInstance struct {
	ServiceId int64
	Port      int32
	// UiPort - the joxit/docker-registry-ui sidecar's own exposed port. Zero
	// is a sentinel meaning "no UI sidecar provisioned yet" (only reachable
	// today through the backfill migration for the pre-existing singleton
	// "registry" plugin) - never treat it as a real port.
	UiPort   int32
	Username string
	// SecretRef - the canonical "scope/owner/key" string form of the
	// domain.SecretRef the password is stored under (see SecretRef.String).
	// Never the value itself - resolved only through
	// internal/service/secrets.Store.
	SecretRef string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UpsertRegistryInstanceReq creates or replaces the velez.registry_instances
// row for a service.
type UpsertRegistryInstanceReq struct {
	ServiceId int64
	Port      int32
	UiPort    int32
	Username  string
	SecretRef string
}

// CreateRegistryInstanceReq is the input to
// ContainerRegistryService.CreateRegistryInstance.
type CreateRegistryInstanceReq struct {
	Name string

	// Environment - the isolated namespace this instance deploys into. Empty
	// means the default/PROD environment (see storage/environments.Resolve).
	Environment string

	// Box - sizing tier override; empty keeps the builtin registry
	// descriptor's own box.
	Box string

	// ExposeToPort - optional host port the instance's registry port (5000)
	// is published on. Zero means the port stays internal to the Docker
	// network.
	ExposeToPort uint32

	// OwnerService - when set, a velez.service_resources binding is recorded
	// (OwnerService, Name, "container_registry") once the instance is
	// created.
	OwnerService string
}

// RegistryInstanceView is one resolved Container-Registry-as-a-Service
// instance: registry-specific facts (RegistryInstance) merged with live
// service/deployment state read through VervServicesService - status,
// environment and image are never persisted alongside the registry-specific
// facts.
type RegistryInstanceView struct {
	Name     string
	Port     int32
	UiPort   int32
	Username string

	Environment string
	Status      string
	ImageName   string

	// OwnerService - unset when the instance was created without one, or
	// when resolving it back from the registry instance's name would need a
	// resource-owner reverse lookup ServiceResourcesStorage doesn't expose
	// today (it only answers "what does service X own", never "who owns
	// resource Y"). Populated at creation time in the immediate
	// CreateRegistryInstance response.
	OwnerService string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ListRegistryInstancesReq pages through every
// Container-Registry-as-a-Service instance.
type ListRegistryInstancesReq struct {
	Paging Paging
}

// RegistryInstanceList is the paged result of
// ContainerRegistryService.ListRegistryInstances. Never carries a password -
// see RegistryInstanceCredentials.
type RegistryInstanceList struct {
	Total     uint64
	Instances []RegistryInstanceView
}

// RegistryInstanceCredentials is the result of ContainerRegistryService.
// GetRegistryInstanceCredentials - the only ContainerRegistry operation that
// resolves a secret_ref to its plaintext value.
type RegistryInstanceCredentials struct {
	Username    string
	Password    string
	RegistryUrl string
}
