package labels

import (
	"github.com/docker/docker/api/types/image"
)

const (
	// CreatedWithVelezLabel - helps Velez identify it's owns containers.
	// Set by default when using docker.Docker
	CreatedWithVelezLabel = "CREATED_WITH_VELEZ"
	Sidecar               = "SIDECAR"
	VervServiceLabel      = "VERV_SERVICE"
	MatreshkaConfigLabel  = "MATRESHKA_CONFIG_ENABLED"
	AutoUpgrade           = "VELEZ_AUTO_UPGRADE"
	// DependsOnLabel — comma-separated list of service names this service depends on.
	// Used by GetServiceGraph to discover and persist service-to-service dependencies.
	DependsOnLabel = "VERV_DEPENDS_ON"
)

func IsMatreshkaImage(r *image.InspectResponse) bool {
	if r == nil || r.Config == nil || len(r.Config.Labels) == 0 {
		return false
	}

	return r.Config.Labels[MatreshkaConfigLabel] == "true"
}
