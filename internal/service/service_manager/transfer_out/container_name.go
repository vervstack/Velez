package transfer_out

import (
	"strings"

	"github.com/docker/docker/api/types/container"
)

func containerName(inspected container.InspectResponse) string {
	return strings.TrimPrefix(inspected.Name, "/")
}

func isRunning(inspected container.InspectResponse) bool {
	return inspected.State != nil && inspected.State.Running
}
