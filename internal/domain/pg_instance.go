package domain

import (
	"time"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// PgInstance is the pg-specific satellite row for a Postgres-as-a-Service
// instance (velez.pg_instances). Status, environment, image, container id and
// deploy history live on the underlying velez.services /
// deployment_specifications / deployments rows and velez.deployments, read
// through VervServicesService and ListSmerds - never duplicated here. See
// docs/features/pgaas_and_registry_plugin.md section 3.
type PgInstance struct {
	ServiceId int64
	DbName    string
	Username  string
	// SecretRef - the canonical "scope/owner/key" string form of the
	// domain.SecretRef the password is stored under (see SecretRef.String).
	// Never the value itself - resolved only through
	// internal/service/secrets.Store.
	SecretRef string
	Port      int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UpsertPgInstanceReq creates or replaces the velez.pg_instances row for a
// service.
type UpsertPgInstanceReq struct {
	ServiceId int64
	DbName    string
	Username  string
	SecretRef string
	Port      int32
}

// CreatePgInstanceReq is the input to PostgresService.CreatePgInstance. See
// docs/features/pgaas_and_registry_plugin.md section 3's create flow.
type CreatePgInstanceReq struct {
	Name string

	// Environment - the isolated namespace this instance deploys into. Empty
	// means the default/PROD environment (see storage/environments.Resolve).
	Environment string

	// Box - sizing tier override; empty keeps the builtin postgres
	// descriptor's own box ("small").
	Box string

	// ExposeToPort - optional host port the instance's 5432 is published on.
	// Zero means the port stays internal to the Docker network.
	ExposeToPort uint32

	// OwnerService - when set, a velez.service_resources binding is recorded
	// (OwnerService, Name, "postgres") once the instance is created.
	OwnerService string

	// Isolation - PostgresAPI.CreatePgInstance's wire request carries no
	// isolation field yet (shared_pool is surfaced UI-disabled-only per the
	// spec's "Isolation" section, with no provisioning path today), so
	// callers through the transport layer always pass
	// PG_INSTANCE_ISOLATION_SEPARATE_INSTANCE here. The field - and the
	// service-layer validation that rejects anything else - exist so the
	// contract is already enforced once a real isolation choice is wired up.
	Isolation velez_api.PgInstanceIsolation
}

// PgInstanceView is one resolved Postgres-as-a-Service instance: pg-specific
// facts (PgInstance) merged with live service/deployment state read through
// VervServicesService - status, environment and image are never persisted
// alongside the pg-specific facts. See section 3's "never duplicated" rule.
type PgInstanceView struct {
	Name     string
	DbName   string
	Username string
	Port     int32

	Environment string
	Status      string
	ImageName   string

	// OwnerService - unset when the instance was created without one, or
	// when resolving it back from the pg instance's name would need a
	// resource-owner reverse lookup ServiceResourcesStorage doesn't expose
	// today (it only answers "what does service X own", never "who owns
	// resource Y"). Populated at creation time in the immediate
	// CreatePgInstance response.
	OwnerService string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ListPgInstancesReq pages through every Postgres-as-a-Service instance.
type ListPgInstancesReq struct {
	Paging Paging
}

// PgInstanceList is the paged result of PostgresService.ListPgInstances.
// Never carries a password - see PgInstanceCredentials.
type PgInstanceList struct {
	Total     uint64
	Instances []PgInstanceView
}

// PgInstanceCredentials is the result of PostgresService
// .GetPgInstanceCredentials - the only PGaaS operation that resolves a
// secret_ref to its plaintext value.
type PgInstanceCredentials struct {
	DbName   string
	Username string
	Password string
	Dsn      string
}
