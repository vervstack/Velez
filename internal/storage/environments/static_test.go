package environments

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
)

func TestStaticStorage_SeedsDefaultEnvironmentWithConfiguredSuffix(t *testing.T) {
	s := NewStatic(nil, "node-a")

	env, err := s.GetEnvironmentByName(context.Background(), DefaultEnvironmentName)
	require.NoError(t, err)
	require.Equal(t, "node-a", env.Suffix)
}

// Configured environment names default their suffix to the name itself, the
// same rule CreateEnvironment applies to an omitted suffix.
func TestStaticStorage_SeedsConfiguredEnvironments(t *testing.T) {
	s := NewStatic([]string{"STAGE", "DEV"}, "")

	list, err := s.ListEnvironments(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 3)

	stage, err := s.GetEnvironmentByName(context.Background(), "STAGE")
	require.NoError(t, err)
	require.Equal(t, "STAGE", stage.Suffix)
}

func TestStaticStorage_CRUDRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := NewStatic(nil, "")

	created, err := s.CreateEnvironment(ctx, domain.CreateEnvironmentReq{Name: "QA", Suffix: "qa"})
	require.NoError(t, err)
	require.NotZero(t, created.ID)

	byID, err := s.GetEnvironmentByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, "QA", byID.Name)

	newSuffix := "qa2"

	updated, err := s.UpdateEnvironment(ctx, domain.UpdateEnvironmentReq{ID: created.ID, Suffix: &newSuffix})
	require.NoError(t, err)
	require.Equal(t, "qa2", updated.Suffix)
	require.Equal(t, "QA", updated.Name, "omitted name must be preserved")

	err = s.DeleteEnvironment(ctx, created.ID)
	require.NoError(t, err)

	_, err = s.GetEnvironmentByID(ctx, created.ID)
	require.True(t, errors.Is(err, storage.ErrNotFound))
}

func TestStaticStorage_CreateDuplicateNameRejected(t *testing.T) {
	ctx := context.Background()
	s := NewStatic(nil, "")

	_, err := s.CreateEnvironment(ctx, domain.CreateEnvironmentReq{Name: DefaultEnvironmentName})
	require.True(t, errors.Is(err, storage.ErrAlreadyExists))
}

func TestStaticStorage_DeleteMissingRejected(t *testing.T) {
	err := NewStatic(nil, "").DeleteEnvironment(context.Background(), 999)
	require.True(t, errors.Is(err, storage.ErrNotFound))
}
