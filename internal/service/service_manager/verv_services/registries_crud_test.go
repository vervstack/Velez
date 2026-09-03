package verv_services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/registries"
)

func newRegistryService(t *testing.T) *VervService {
	t.Helper()

	dataStorage := &testStorage{
		registries: registries.NewStatic(),
	}

	return New(dataStorage, nil, nil, nil, nil, nil)
}

// CreateRegistry with IsDefault=false never touches the transaction manager -
// testStorage.txManager is left nil (its zero value), so this test would
// panic on a nil pointer dereference if it did.
func TestVervService_CreateRegistry_NonDefaultSkipsTransaction(t *testing.T) {
	v := newRegistryService(t)

	req := domain.CreateRegistryReq{
		Name: "docker-hub",
		Type: domain.RegistryTypeDockerHub,
	}

	reg, err := v.CreateRegistry(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "docker-hub", reg.Name)
	require.False(t, reg.IsDefault)
}

func TestVervService_CreateRegistry_EmptyNameRejected(t *testing.T) {
	v := newRegistryService(t)

	_, err := v.CreateRegistry(context.Background(), domain.CreateRegistryReq{Type: domain.RegistryTypeDockerHub})
	require.ErrorIs(t, err, ErrRegistryNameRequired)
}

func TestVervService_CreateRegistry_InvalidTypeRejected(t *testing.T) {
	v := newRegistryService(t)

	req := domain.CreateRegistryReq{Name: "bad", Type: domain.RegistryType("bogus")}

	_, err := v.CreateRegistry(context.Background(), req)
	require.ErrorIs(t, err, ErrInvalidRegistryType)
}

func TestVervService_UpdateRegistry_MissingIDRejected(t *testing.T) {
	v := newRegistryService(t)

	_, err := v.UpdateRegistry(context.Background(), domain.UpdateRegistryReq{})
	require.Error(t, err)
}

// UpdateRegistry with IsDefault omitted never touches the transaction
// manager, mirroring CreateRegistry's non-default path.
func TestVervService_UpdateRegistry_NonDefaultSkipsTransaction(t *testing.T) {
	ctx := context.Background()
	v := newRegistryService(t)

	created, err := v.CreateRegistry(ctx, domain.CreateRegistryReq{
		Name: "docker-hub",
		Type: domain.RegistryTypeDockerHub,
	})
	require.NoError(t, err)

	newName := "docker-hub-renamed"

	updated, err := v.UpdateRegistry(ctx, domain.UpdateRegistryReq{
		ID:   created.ID,
		Name: &newName,
	})
	require.NoError(t, err)
	require.Equal(t, newName, updated.Name)
}

func TestVervService_DeleteRegistry_RequiresIDOrName(t *testing.T) {
	v := newRegistryService(t)

	err := v.DeleteRegistry(context.Background(), domain.DeleteRegistryReq{})
	require.Error(t, err)
}

func TestVervService_DeleteRegistry_UnknownNameRejected(t *testing.T) {
	v := newRegistryService(t)

	name := "ghost"

	err := v.DeleteRegistry(context.Background(), domain.DeleteRegistryReq{Name: &name})
	require.ErrorIs(t, err, ErrRegistryNotFound)
}

func TestVervService_GetRegistry_UnknownIDRejected(t *testing.T) {
	v := newRegistryService(t)

	_, err := v.GetRegistry(context.Background(), 999)
	require.ErrorIs(t, err, ErrRegistryNotFound)
}
