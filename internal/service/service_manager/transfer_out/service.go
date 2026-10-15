// Package transfer_out implements service.TransferOutService - the source side
// of a container transfer.
package transfer_out

import (
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/transfer_io"
	"go.vervstack.ru/Velez/internal/storage"
)

//nolint:unused // dependencies are consumed once the stubs are filled in
type Service struct {
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
) *Service {
	return &Service{
		dataStorage: dataStorage,
		runtimes:    runtimes,
		transferIO:  transferIO,
		docker:      docker,
	}
}
