package registries

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func TestStaticStorage_SeedsBuiltinDockerHubRegistry(t *testing.T) {
	s := NewStatic()

	list, err := s.ListRegistries(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, dockerHubRegistryName, list[0].Name)
	require.Equal(t, domain.RegistryTypeDockerHub, list[0].Type)
	require.True(t, list[0].IsDefault)
}

func TestStaticStorage_WriteMethodsRequireStatefullMode(t *testing.T) {
	ctx := context.Background()
	s := NewStatic()

	_, err := s.CreateRegistry(ctx, domain.CreateRegistryReq{Name: "my-registry"})
	require.True(t, errors.Is(err, user_errors.ErrRequiresStatefullMode))

	newUrl := "https://example.com"

	_, err = s.UpdateRegistry(ctx, domain.UpdateRegistryReq{ID: 1, Url: &newUrl})
	require.True(t, errors.Is(err, user_errors.ErrRequiresStatefullMode))

	err = s.DeleteRegistry(ctx, 1)
	require.True(t, errors.Is(err, user_errors.ErrRequiresStatefullMode))

	err = s.ClearDefaultRegistry(ctx)
	require.True(t, errors.Is(err, user_errors.ErrRequiresStatefullMode))
}
