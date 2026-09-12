package verv_services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/local_storage"
	"go.vervstack.ru/Velez/internal/storage/registries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testRegistryName = "docker-hub"
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
		Name: testRegistryName,
		Type: domain.RegistryTypeDockerHub,
	}

	_, err := v.CreateRegistry(context.Background(), req)
	require.ErrorIs(t, err, user_errors.ErrRequiresStatefullMode)
}

func TestVervService_CreateRegistry_EmptyNameRejected(t *testing.T) {
	v := newRegistryService(t)

	_, err := v.CreateRegistry(context.Background(), domain.CreateRegistryReq{Type: domain.RegistryTypeDockerHub})
	require.ErrorIs(t, err, user_errors.ErrRegistryNameRequired)
}

func TestVervService_CreateRegistry_InvalidTypeRejected(t *testing.T) {
	v := newRegistryService(t)

	req := domain.CreateRegistryReq{Name: "bad", Type: domain.RegistryType("bogus")}

	_, err := v.CreateRegistry(context.Background(), req)
	require.ErrorIs(t, err, user_errors.ErrInvalidRegistryType)
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

	newName := "docker-hub-renamed"

	_, err := v.UpdateRegistry(ctx, domain.UpdateRegistryReq{
		ID:   1,
		Name: &newName,
	})
	require.ErrorIs(t, err, user_errors.ErrRequiresStatefullMode)
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
	require.ErrorIs(t, err, user_errors.ErrRegistryNotFound)
}

func TestVervService_GetRegistry_UnknownIDRejected(t *testing.T) {
	v := newRegistryService(t)

	_, err := v.GetRegistry(context.Background(), 999)
	require.ErrorIs(t, err, user_errors.ErrRegistryNotFound)
}

// newLocalStorageRegistryService builds a VervService against a real
// local_storage.New(...)-backed storage.Storage - single-node/dev mode,
// exactly as internal/app/custom.go wires it when no Postgres is configured.
// Unlike newRegistryService's hand-rolled testStorage fake, this exercises
// the real localStorage.TxManager() / deployments.Execute path.
func newLocalStorageRegistryService(t *testing.T) *VervService {
	t.Helper()

	dataStorage := local_storage.New(nil, config.Config{})

	return New(dataStorage, nil, nil, nil, nil, nil)
}

// TestVervService_CreateRegistry_DefaultAgainstLocalStorage reproduces the
// panic CreateRegistry(IsDefault: true) hit against local_storage before
// localStorage.TxManager() returned a real Transactor instead of nil:
// v.dataStorage.TxManager().Execute(...) dereferenced a nil *sqldb.TxManager.
// With TxManager() always non-nil, this runs cleanly and surfaces
// user_errors.ErrRequiresStatefullMode - single-node/dev mode's registries
// storage rejects every write, IsDefault ones included.
func TestVervService_CreateRegistry_DefaultAgainstLocalStorage(t *testing.T) {
	v := newLocalStorageRegistryService(t)

	req := domain.CreateRegistryReq{
		Name:      testRegistryName,
		Type:      domain.RegistryTypeDockerHub,
		IsDefault: true,
	}

	_, err := v.CreateRegistry(context.Background(), req)
	require.ErrorIs(t, err, user_errors.ErrRequiresStatefullMode)
}

// TestVervService_UpdateRegistry_DefaultAgainstLocalStorage is
// UpdateRegistry's counterpart to
// TestVervService_CreateRegistry_DefaultAgainstLocalStorage - its IsDefault:
// true path hits the same TxManager().Execute(...) call.
func TestVervService_UpdateRegistry_DefaultAgainstLocalStorage(t *testing.T) {
	ctx := context.Background()
	v := newLocalStorageRegistryService(t)

	isDefault := true

	_, err := v.UpdateRegistry(ctx, domain.UpdateRegistryReq{
		ID:        1,
		IsDefault: &isDefault,
	})
	require.ErrorIs(t, err, user_errors.ErrRequiresStatefullMode)
}
