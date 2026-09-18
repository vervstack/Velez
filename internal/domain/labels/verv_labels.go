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

	// RunnerInstanceLabel marks a container as a Runners-as-a-Service instance
	// provisioned by runneraas.CreateRunner. RunnerProviderLabel/
	// RunnerScopeLabel/RunnerTargetLabel/RunnerLabelsLabel carry the runner's
	// facts directly as container labels rather than provider-specific env
	// vars - the single-node/dev local_storage backend reads them straight off
	// the container (see internal/storage/local_storage/runners.go), which is
	// what keeps that backend provider-agnostic instead of having to parse
	// each Provider's own env var vocabulary. Always paired with
	// VervServiceLabel set to the instance name. Always inert in cluster mode,
	// where velez.runners is authoritative.
	RunnerInstanceLabel = "velez.runner"
	RunnerProviderLabel = "velez.runner.provider"
	RunnerScopeLabel    = "velez.runner.scope"
	RunnerTargetLabel   = "velez.runner.target"
	RunnerLabelsLabel   = "velez.runner.labels"
	// RunnerBaseUrlLabel carries a runner's git provider base URL (e.g. a
	// self-managed GitLab instance's URL). Empty for a provider with a
	// single fixed API host (GitHub).
	RunnerBaseUrlLabel = "velez.runner.base_url"

	// RegistryaasInstanceLabel marks a container as a
	// Container-Registry-as-a-Service instance provisioned by
	// registryaas.CreateRegistryInstance. Credentials live in an htpasswd
	// file inside the registry-auth volume, not container env vars, so -
	// unlike PgaasInstanceLabel - username, port and ui_port are carried
	// directly as labels for the single-node/dev local_storage backend to
	// read back (see internal/storage/local_storage/registry_instances.go),
	// the same label-derived approach RunnerInstanceLabel uses. Always
	// paired with VervServiceLabel set to the instance name. Always inert in
	// cluster mode, where velez.registry_instances is authoritative.
	RegistryaasInstanceLabel = "velez.registryaas"
	RegistryaasUsernameLabel = "velez.registryaas.username"
	RegistryaasPortLabel     = "velez.registryaas.port"
	RegistryaasUiPortLabel   = "velez.registryaas.ui_port"

	// PgaasNamePrefix, RegistryaasNamePrefix, GitlabRunnerNamePrefix and
	// GithubRunnerNamePrefix are prepended to the instance name at the point
	// each service type first derives it from the request - see
	// pgaas.buildPgDescriptor, create_registry_instance.go's BuildJobs and
	// create_runner.go's BuildJobs. Applying the prefix there, rather than at
	// the shared create_smerd container-create chokepoint, keeps the
	// container name the same string used everywhere else to look the
	// instance back up (volume name, hostname, VervServiceLabel, DB/list join
	// key).
	PgaasNamePrefix        = "pgaas_"
	RegistryaasNamePrefix  = "cr_"
	GitlabRunnerNamePrefix = "gitlab_runner_"
	GithubRunnerNamePrefix = "github_runner_"

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
