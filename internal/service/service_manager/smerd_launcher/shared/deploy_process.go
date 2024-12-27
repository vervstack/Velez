package shared

import (
	"github.com/docker/docker/api/types"
)

type DeployProcess struct {
	Image     types.ImageInspect
	Container *types.ContainerJSON
}
