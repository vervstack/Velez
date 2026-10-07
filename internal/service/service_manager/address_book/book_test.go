package address_book

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/makosh/pkg/registry"
)

const (
	testDockerPort = 32001
	testBookHost   = "node.local"
)

func newTestBook() *Book {
	client := registry.NewClient(registry.New(), "test")

	return New(nil, nil, nil, nil, client)
}

func newDockerAddress(host string, port uint32) domain.ServiceAddress {
	return domain.ServiceAddress{
		ServiceName: testService,
		Name:        domain.WebUiAddressName,
		Host:        host,
		Port:        port,
		Scope:       velez_api.AddressScope_ADDRESS_SCOPE_DOCKER,
	}
}

func newVcnAddress(host string, port uint32) domain.ServiceAddress {
	address := newDockerAddress(host, port)

	address.Scope = velez_api.AddressScope_ADDRESS_SCOPE_VCN

	return address
}

func Test_Addresses_UnknownServiceIsEmpty(t *testing.T) {
	book := newTestBook()

	got, err := book.Addresses(context.Background(), "missing", domain.WebUiAddressName)

	require.NoError(t, err)
	require.Empty(t, got)
}

func Test_Store_RoundTripsBothScopesDockerFirst(t *testing.T) {
	ctx := context.Background()
	book := newTestBook()

	found := serviceAddresses{
		newVcnAddress(testVcnIpV4, testUiPort),
		newDockerAddress(testBookHost, testDockerPort),
	}

	err := book.store(ctx, testService, found)
	require.NoError(t, err)

	got, err := book.Addresses(ctx, testService, domain.WebUiAddressName)
	require.NoError(t, err)

	want := []domain.ServiceAddress{
		newDockerAddress(testBookHost, testDockerPort),
		newVcnAddress(testVcnIpV4, testUiPort),
	}

	require.Equal(t, want, got)
}

func Test_Store_EmptyHostRoundTrips(t *testing.T) {
	ctx := context.Background()
	book := newTestBook()

	err := book.store(ctx, testService, serviceAddresses{newDockerAddress("", 9000)})
	require.NoError(t, err)

	got, err := book.Addresses(ctx, testService, domain.WebUiAddressName)
	require.NoError(t, err)
	require.Equal(t, []domain.ServiceAddress{newDockerAddress("", 9000)}, got)
}

func Test_Store_MissingScopeIsDeleted(t *testing.T) {
	ctx := context.Background()
	book := newTestBook()

	both := serviceAddresses{
		newDockerAddress(testBookHost, testDockerPort),
		newVcnAddress(testVcnIpV4, testUiPort),
	}

	err := book.store(ctx, testService, both)
	require.NoError(t, err)

	onlyDocker := serviceAddresses{newDockerAddress(testBookHost, testDockerPort)}

	err = book.store(ctx, testService, onlyDocker)
	require.NoError(t, err)

	got, err := book.Addresses(ctx, testService, domain.WebUiAddressName)
	require.NoError(t, err)
	require.Equal(t, []domain.ServiceAddress{newDockerAddress(testBookHost, testDockerPort)}, got)
}

func Test_Store_EmptyDeletesBothScopes(t *testing.T) {
	ctx := context.Background()
	book := newTestBook()

	both := serviceAddresses{
		newDockerAddress(testBookHost, testDockerPort),
		newVcnAddress(testVcnIpV4, testUiPort),
	}

	err := book.store(ctx, testService, both)
	require.NoError(t, err)

	err = book.store(ctx, testService, nil)
	require.NoError(t, err)

	got, err := book.Addresses(ctx, testService, domain.WebUiAddressName)
	require.NoError(t, err)
	require.Empty(t, got)
}

func Test_Store_UnknownServiceEmptyDoesNotFail(t *testing.T) {
	book := newTestBook()

	err := book.store(context.Background(), "missing", nil)

	require.NoError(t, err)
}

func Test_Addresses_FiltersByName(t *testing.T) {
	ctx := context.Background()
	book := newTestBook()

	err := book.store(ctx, testService, serviceAddresses{newDockerAddress(testBookHost, testDockerPort)})
	require.NoError(t, err)

	got, err := book.Addresses(ctx, testService, "api")

	require.NoError(t, err)
	require.Empty(t, got)
}

func Test_Drop_RemovesBothScopesAndForgets(t *testing.T) {
	ctx := context.Background()
	book := newTestBook()

	both := serviceAddresses{
		newDockerAddress(testBookHost, testDockerPort),
		newVcnAddress(testVcnIpV4, testUiPort),
	}

	err := book.store(ctx, testService, both)
	require.NoError(t, err)

	book.remember(testService)

	err = book.Drop(ctx, testService)
	require.NoError(t, err)

	got, err := book.Addresses(ctx, testService, domain.WebUiAddressName)
	require.NoError(t, err)
	require.Empty(t, got)
	require.Empty(t, book.known)
}

func Test_Drop_UnknownServiceDoesNotFail(t *testing.T) {
	book := newTestBook()

	err := book.Drop(context.Background(), "missing")

	require.NoError(t, err)
}

func Test_ReplaceKnown_ReturnsStaleNames(t *testing.T) {
	book := newTestBook()
	book.remember("gone")
	book.remember(testService)

	stale := book.replaceKnown(map[string]serviceAddresses{testService: nil, "fresh": nil})

	require.Equal(t, []string{"gone"}, stale)
	require.Len(t, book.known, 2)
	require.Contains(t, book.known, "fresh")
}
