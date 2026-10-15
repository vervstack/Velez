package transfer_in

import (
	"github.com/google/uuid"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
)

// receiveRun is the state of one incoming transfer; everything it records is
// what a failed run has to remove again.
type receiveRun struct {
	transferUuid        uuid.UUID
	manifest            transferManifest
	runtime             container_runtime.ContainerRuntime
	createdContainerIds []string
	createdVolumes      []string
}
