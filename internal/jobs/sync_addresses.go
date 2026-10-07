package jobs

import (
	"context"

	"github.com/rs/zerolog/log"

	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/service_manager/address_book"
)

const (
	stepSyncAddresses = "sync_addresses"
)

type rootServiceAccessor interface {
	RootService() string
}

// smerdRootService resolves the service a created container's addresses are
// registered under from the labels prepare_verv_config left on the request.
type smerdRootService struct {
	req smerdRequestAccessor
}

func (s smerdRootService) RootService() string {
	return address_book.RootServiceName(s.req.GetRequest().GetLabels())
}

type staticRootService string

func (s staticRootService) RootService() string {
	return string(s)
}

// syncAddressesJob is best-effort: a registry failure is logged and never
// fails or retries the task.
type syncAddressesJob struct {
	addressBook service.AddressBook
	root        rootServiceAccessor
}

func (j *syncAddressesJob) Do(ctx context.Context) error {
	syncAddresses(ctx, j.addressBook, j.root.RootService())

	return nil
}

func syncAddresses(ctx context.Context, addressBook service.AddressBook, rootService string) {
	if rootService == "" {
		return
	}

	err := addressBook.Sync(ctx, rootService)
	if err != nil {
		log.Ctx(ctx).Warn().
			Err(err).
			Str("service", rootService).
			Msg("error syncing service addresses")
	}
}
