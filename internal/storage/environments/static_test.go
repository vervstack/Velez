package environments

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
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

func TestStaticStorage_WriteMethodsRequireStatefullMode(t *testing.T) {
	ctx := context.Background()
	s := NewStatic(nil, "")

	_, err := s.CreateEnvironment(ctx, domain.CreateEnvironmentReq{Name: "QA", Suffix: "qa"})
	require.True(t, errors.Is(err, user_errors.ErrRequiresStatefullMode))

	newSuffix := "qa2"

	_, err = s.UpdateEnvironment(ctx, domain.UpdateEnvironmentReq{ID: 1, Suffix: &newSuffix})
	require.True(t, errors.Is(err, user_errors.ErrRequiresStatefullMode))

	err = s.DeleteEnvironment(ctx, 1)
	require.True(t, errors.Is(err, user_errors.ErrRequiresStatefullMode))
}
