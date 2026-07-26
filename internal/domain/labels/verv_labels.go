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
)

func IsMatreshkaImage(r *image.InspectResponse) bool {
	if r == nil || r.Config == nil || len(r.Config.Labels) == 0 {
		return false
	}

	return r.Config.Labels[MatreshkaConfigLabel] == "true"
}
