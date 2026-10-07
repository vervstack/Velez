package address_book

import (
	"context"
	"errors"
	"net"
	"sort"
	"strconv"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	makosh "go.vervstack.ru/makosh/pkg/makosh_be"
)

var registryScopes = []velez_api.AddressScope{
	velez_api.AddressScope_ADDRESS_SCOPE_DOCKER,
	velez_api.AddressScope_ADDRESS_SCOPE_VCN,
}

func toMakoshScope(scope velez_api.AddressScope) makosh.AddressScope {
	switch scope {
	case velez_api.AddressScope_ADDRESS_SCOPE_DOCKER:
		return makosh.AddressScope_DOCKER
	case velez_api.AddressScope_ADDRESS_SCOPE_VCN:
		return makosh.AddressScope_VCN
	default:
		return makosh.AddressScope_ADDRESS_SCOPE_UNSPECIFIED
	}
}

func fromMakoshScope(scope makosh.AddressScope) velez_api.AddressScope {
	switch scope {
	case makosh.AddressScope_DOCKER:
		return velez_api.AddressScope_ADDRESS_SCOPE_DOCKER
	case makosh.AddressScope_VCN:
		return velez_api.AddressScope_ADDRESS_SCOPE_VCN
	default:
		return velez_api.AddressScope_ADDRESS_SCOPE_UNSPECIFIED
	}
}

func (b *Book) store(ctx context.Context, serviceName string, found serviceAddresses) error {
	var storeErrs []error

	for _, scope := range registryScopes {
		var scoped []*makosh.Address

		for _, address := range found {
			if address.Scope != scope {
				continue
			}

			scoped = append(scoped, &makosh.Address{
				Addr:  net.JoinHostPort(address.Host, strconv.FormatUint(uint64(address.Port), 10)),
				Scope: toMakoshScope(scope),
				Name:  address.Name,
			})
		}

		var err error

		if len(scoped) > 0 {
			err = b.upsertScope(ctx, serviceName, scoped)
		} else {
			err = b.deleteScope(ctx, serviceName, scope)
		}

		if err != nil {
			storeErrs = append(storeErrs, err)
		}
	}

	return errors.Join(storeErrs...)
}

func (b *Book) upsertScope(ctx context.Context, serviceName string, addresses []*makosh.Address) error {
	endpoint := &makosh.Endpoint{
		ServiceName: serviceName,
		Addresses:   addresses,
	}

	req := &makosh.UpsertEndpoints_Request{
		Endpoints: []*makosh.Endpoint{endpoint},
	}

	_, err := b.registry.UpsertEndpoints(ctx, req)
	if err != nil {
		return rerrors.Wrap(err, "error upserting service addresses")
	}

	return nil
}

func (b *Book) deleteScope(ctx context.Context, serviceName string, scope velez_api.AddressScope) error {
	req := &makosh.DeleteEndpoints_Request{
		ServiceName: serviceName,
		Scope:       toMakoshScope(scope),
	}

	_, err := b.registry.DeleteEndpoints(ctx, req)
	if status.Code(err) == codes.NotFound {
		log.Ctx(ctx).Debug().
			Str("service", serviceName).
			Msg("no addresses to delete")

		return nil
	}

	if err != nil {
		return rerrors.Wrap(err, "error deleting service addresses")
	}

	return nil
}

func (b *Book) Addresses(ctx context.Context, serviceName, addressName string) ([]domain.ServiceAddress, error) {
	req := &makosh.ListEndpoints_Request{
		ServiceName: serviceName,
		Scope:       makosh.AddressScope_ADDRESS_SCOPE_UNSPECIFIED,
	}

	resp, err := b.registry.ListEndpoints(ctx, req)
	if status.Code(err) == codes.NotFound {
		return []domain.ServiceAddress{}, nil
	}

	if err != nil {
		return nil, rerrors.Wrap(err, "error listing service addresses")
	}

	addresses := make([]domain.ServiceAddress, 0, len(resp.GetAddresses()))

	for _, stored := range resp.GetAddresses() {
		if stored.GetName() != addressName {
			continue
		}

		host, rawPort, splitErr := net.SplitHostPort(stored.GetAddr())
		if splitErr != nil {
			continue
		}

		port, parseErr := strconv.ParseUint(rawPort, 10, 32)
		if parseErr != nil {
			continue
		}

		address := domain.ServiceAddress{
			ServiceName: serviceName,
			Name:        addressName,
			Host:        host,
			Port:        uint32(port),
			Scope:       fromMakoshScope(stored.GetScope()),
		}

		addresses = append(addresses, address)
	}

	sort.SliceStable(addresses, func(i, j int) bool {
		return addresses[i].Scope < addresses[j].Scope
	})

	return addresses, nil
}

func (b *Book) Drop(ctx context.Context, serviceName string) error {
	var dropErrs []error

	for _, scope := range registryScopes {
		err := b.deleteScope(ctx, serviceName, scope)
		if err != nil {
			dropErrs = append(dropErrs, err)
		}
	}

	b.forget(serviceName)

	err := errors.Join(dropErrs...)
	if err != nil {
		return rerrors.Wrap(err, "error dropping service addresses")
	}

	return nil
}
