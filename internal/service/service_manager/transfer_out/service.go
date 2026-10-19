// Package transfer_out implements service.TransferOutService - the source side
// of a container transfer.
package transfer_out

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/transfer_io"
	"go.vervstack.ru/Velez/internal/storage"
)

type TransferOutServiceImpl struct {
	dataStorage storage.Storage
	runtimes    container_runtime.RuntimeResolver
	transferIO  transfer_io.TransferIO
	docker      node_clients.Docker

	mu      sync.Mutex
	cancels map[uuid.UUID]context.CancelFunc
}

func New(
	dataStorage storage.Storage,
	runtimes container_runtime.RuntimeResolver,
	transferIO transfer_io.TransferIO,
	docker node_clients.Docker,
) *TransferOutServiceImpl {
	return &TransferOutServiceImpl{
		dataStorage: dataStorage,
		runtimes:    runtimes,
		transferIO:  transferIO,
		docker:      docker,
		cancels:     make(map[uuid.UUID]context.CancelFunc),
	}
}
