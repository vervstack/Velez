package domain

import (
	"github.com/docker/docker/api/types"

	"github.com/godverv/Velez/pkg/velez_api"
)

type LaunchSmerd struct {
	*velez_api.CreateSmerd_Request
}

type LaunchSmerdState struct {
	Image       types.ImageInspect
	ContainerId *string
}

type LaunchSmerdResult struct {
	ContainerId string
}
