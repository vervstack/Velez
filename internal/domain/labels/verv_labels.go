package labels

import (
	"github.com/docker/docker/api/types/image"
)

const (
	CreatedWithVelezLabel = "CREATED_WITH_VELEZ"
	MatreshkaConfigLabel  = "MATRESHKA_CONFIG_ENABLED"
)

func IsMatreshkaImage(r *image.InspectResponse) bool {
	if r == nil || r.Config == nil || len(r.Config.Labels) == 0 {
		return false
	}

	return r.Config.Labels[MatreshkaConfigLabel] == "true"
}
