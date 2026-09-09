package labels

import (
	"github.com/docker/docker/api/types/image"
)

const (
	// CreatedWithVelezLabel - helps Velez identify it's owns containers.
	// Set by default when using docker.Docker.
	CreatedWithVelezLabel = "CREATED_WITH_VELEZ"
	Sidecar               = "SIDECAR"
	VervServiceLabel      = "VERV_SERVICE"
	MatreshkaConfigLabel  = "MATRESHKA_CONFIG_ENABLED"
	AutoUpgrade           = "VELEZ_AUTO_UPGRADE"
	// DependsOnLabel — comma-separated list of service names this service depends on.
	// Used by GetServiceGraph to discover and persist service-to-service dependencies.
	DependsOnLabel = "VERV_DEPENDS_ON"
	// SuffixLabel stores the ContainerSuffix value this Velez instance was configured with.
	// Used to isolate containers when multiple Velez instances share the same Docker host.
	SuffixLabel = "VELEZ_SUFFIX"

	// Service metadata labels — stored on Docker containers, surfaced via AboutService.
	DescriptionLabel = "velez.description"
	ServiceTypeLabel = "velez.type"
	TeamLabel        = "velez.team"
	RepoLabel        = "velez.repo"
	PortLabel        = "velez.port"
	EnvLabel         = "env"

	// PgaasInstanceLabel marks a container as a Postgres-as-a-Service instance
	// provisioned by pgaas.CreatePgInstance. In single-node/dev mode (no
	// velez.pg_instances / velez.secrets tables) the local_storage backend
	// treats the labelled container as the system of record for the
	// instance's db name, user, port and generated password - see
	// internal/storage/local_storage/pg_instances.go and secrets.go. Always
	// paired with VervServiceLabel set to the instance name.
	PgaasInstanceLabel = "velez.pgaas"

	// TagLabelPrefix - per docs/features/vervonomicon.md's "Mapping onto
	// CreateSmerd.Request" table, each vervonomicon service.tags entry
	// becomes a label "verv.tag.<tag>", mirroring the dotted velez.*
	// metadata labels above rather than the older SCREAMING_SNAKE ones.
	TagLabelPrefix = "verv.tag."

	// TagLabelValue is the value written for every verv.tag.<tag> label -
	// the tag's presence is the signal, matching the boolean-label
	// convention container_manager already uses ("true").
	TagLabelValue = "true"
)

func IsMatreshkaImage(r *image.InspectResponse) bool {
	if r == nil || r.Config == nil || len(r.Config.Labels) == 0 {
		return false
	}

	return r.Config.Labels[MatreshkaConfigLabel] == "true"
}
