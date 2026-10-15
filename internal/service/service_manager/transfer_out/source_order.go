package transfer_out

import (
	"github.com/docker/docker/api/types/container"
)

// sourceOrder names the source containers that were running: stopping goes
// sidecars first, restoring goes root first.
type sourceOrder struct {
	stop    []string
	restore []string
}

func newSourceOrder(root container.InspectResponse, sidecars []container.InspectResponse) sourceOrder {
	running := make([]string, 0, len(sidecars))

	for _, sidecar := range sidecars {
		if isRunning(sidecar) {
			running = append(running, containerName(sidecar))
		}
	}

	order := sourceOrder{stop: append([]string{}, running...)}

	if isRunning(root) {
		order.stop = append(order.stop, containerName(root))
		order.restore = append(order.restore, containerName(root))
	}

	order.restore = append(order.restore, running...)

	return order
}
