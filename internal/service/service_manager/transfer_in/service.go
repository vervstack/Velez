// Package transfer_in implements service.TransferInService - the target side
// of a container transfer.
package transfer_in

import (
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/transfer_io"
	"go.vervstack.ru/Velez/internal/storage"
)

type TransferInServiceImpl struct {
	dataStorage storage.Storage
	runtimes    container_runtime.RuntimeResolver
	transferIO  transfer_io.TransferIO
	docker      node_clients.Docker
}

func New(
	dataStorage storage.Storage,
	runtimes container_runtime.RuntimeResolver,
	transferIO transfer_io.TransferIO,
	docker node_clients.Docker,
) *TransferInServiceImpl {
	return &TransferInServiceImpl{
		dataStorage: dataStorage,
		runtimes:    runtimes,
		transferIO:  transferIO,
		docker:      docker,
	}
}
