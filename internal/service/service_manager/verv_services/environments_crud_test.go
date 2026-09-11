package verv_services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/environments"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testEnvStage = "STAGE"
)

func newEnvService(t *testing.T, seedNames []string, defaultSuffix string) *VervService {
	t.Helper()

	dataStorage := &testStorage{
		environments: environments.NewStatic(seedNames, defaultSuffix),
	}

	return New(dataStorage, nil, nil, nil, nil, nil)
}

func TestVervService_ListEnvironments(t *testing.T) {
	v := newEnvService(t, []string{testEnvStage}, "prod")

	envs, err := v.ListEnvironments(context.Background())
	require.NoError(t, err)
	require.Len(t, envs, 2)
	require.Equal(t, environments.DefaultEnvironmentName, envs[0].Name)
	require.Equal(t, "prod", envs[0].Suffix)
}

// Single-node/dev mode's environments storage is a fixed, baked-in set -
// CreateEnvironment always fails there, an omitted suffix included.
func TestVervService_CreateEnvironment_RequiresStatefullMode(t *testing.T) {
	v := newEnvService(t, nil, "")

	req := domain.CreateEnvironmentReq{Name: testEnvStage}

	_, err := v.CreateEnvironment(context.Background(), req)
	require.ErrorIs(t, err, user_errors.ErrRequiresStatefullMode)
}

func TestVervService_CreateEnvironment_EmptyNameRejected(t *testing.T) {
	v := newEnvService(t, nil, "")

	_, err := v.CreateEnvironment(context.Background(), domain.CreateEnvironmentReq{})
	require.ErrorIs(t, err, user_errors.ErrEnvironmentRequired)
}

func TestVervService_UpdateEnvironment_RequiresStatefullMode(t *testing.T) {
	ctx := context.Background()
	v := newEnvService(t, nil, "")

	newName := "STAGING"
	newSuffix := "stg"

	_, err := v.UpdateEnvironment(ctx, domain.UpdateEnvironmentReq{
		ID:     1,
		Name:   &newName,
		Suffix: &newSuffix,
	})
	require.ErrorIs(t, err, user_errors.ErrRequiresStatefullMode)
}

func TestVervService_UpdateEnvironment_MissingIDRejected(t *testing.T) {
	v := newEnvService(t, nil, "")

	_, err := v.UpdateEnvironment(context.Background(), domain.UpdateEnvironmentReq{})
	require.Error(t, err)
}

// Suffix resolution is the validation entry point for the six request messages
// that carry an `environment` field.
func TestVervService_ResolveEnvironmentSuffix(t *testing.T) {
	v := newEnvService(t, nil, "prod")

	suffix, err := v.ResolveEnvironmentSuffix(context.Background(), environments.DefaultEnvironmentName)
	require.NoError(t, err)
	require.Equal(t, "prod", suffix)
}

// Velez must keep working as a single-environment node: an omitted environment
// is NOT an error, it silently resolves to the default environment's suffix -
// i.e. whatever ContainerSuffix this node was configured with before
// environments existed.
func TestVervService_ResolveEnvironmentSuffix_EmptyDefaultsToDefaultEnvironment(t *testing.T) {
	v := newEnvService(t, nil, "prod")

	suffix, err := v.ResolveEnvironmentSuffix(context.Background(), "")
	require.NoError(t, err)
	require.Equal(t, "prod", suffix)
}

// The common single-node case: ContainerSuffix was never configured, so the
// default environment's suffix is empty and an environment-less request keeps
// producing byte-for-byte the container names/labels it always did.
func TestVervService_ResolveEnvironmentSuffix_EmptyDefaultsToUnsuffixedNode(t *testing.T) {
	v := newEnvService(t, nil, "")

	suffix, err := v.ResolveEnvironmentSuffix(context.Background(), "")
	require.NoError(t, err)
	require.Empty(t, suffix)
}

// The empty name resolves to the DEFAULT environment, not to "the first
// environment that happens to exist" - an explicitly named one keeps its own,
// distinct suffix.
func TestVervService_ResolveEnvironmentSuffix_ExplicitEnvironmentGetsOwnSuffix(t *testing.T) {
	ctx := context.Background()
	v := newEnvService(t, []string{testEnvStage}, "prod")

	defaultSuffix, err := v.ResolveEnvironmentSuffix(ctx, "")
	require.NoError(t, err)

	stageSuffix, err := v.ResolveEnvironmentSuffix(ctx, testEnvStage)
	require.NoError(t, err)

	require.Equal(t, "prod", defaultSuffix)
	require.Equal(t, testEnvStage, stageSuffix)
	require.NotEqual(t, defaultSuffix, stageSuffix)
}

// GetEnvironment is the "look up THIS named environment" primitive and stays
// strict - only ResolveEnvironmentSuffix defaults.
func TestVervService_GetEnvironment_EmptyStillRejected(t *testing.T) {
	v := newEnvService(t, nil, "prod")

	_, err := v.GetEnvironment(context.Background(), "")
	require.ErrorIs(t, err, user_errors.ErrEnvironmentRequired)
}

func TestVervService_ResolveEnvironmentSuffix_UnknownRejected(t *testing.T) {
	v := newEnvService(t, nil, "prod")

	_, err := v.ResolveEnvironmentSuffix(context.Background(), "does-not-exist")
	require.ErrorIs(t, err, user_errors.ErrEnvironmentNotFound)
}

func TestVervService_DeleteEnvironment_RequiresIDOrName(t *testing.T) {
	v := newEnvService(t, nil, "prod")

	err := v.DeleteEnvironment(context.Background(), domain.DeleteEnvironmentReq{})
	require.Error(t, err)
}

func TestVervService_DeleteEnvironment_UnknownNameRejected(t *testing.T) {
	v := newEnvService(t, nil, "prod")

	name := "ghost"

	err := v.DeleteEnvironment(context.Background(), domain.DeleteEnvironmentReq{Name: &name})
	require.ErrorIs(t, err, user_errors.ErrEnvironmentNotFound)
}
