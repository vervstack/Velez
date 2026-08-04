package container_runtime

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

const (
	testProdSuffix = "prod"
	testStageEnv   = "STAGE"
)

type staticProvider struct {
	storage storage.EnvironmentsStorage
}

func (p *staticProvider) Environments() storage.EnvironmentsStorage {
	return p.storage
}

// dedicatedStorage yields a single environment bound to a dedicated Docker
// host. Nothing writes domain.Environment.DockerHost yet (Phase 2), so the
// resolver's dedicated branch needs a storage that can produce one.
type dedicatedStorage struct {
	storage.EnvironmentsStorage

	env domain.Environment
}

func (s *dedicatedStorage) GetEnvironmentByName(_ context.Context, name string) (domain.Environment, error) {
	if name != s.env.Name {
		return domain.Environment{}, storage.ErrNotFound
	}

	return s.env, nil
}

func TestResolver_ResolvesLabelBasedRuntimeWithEnvironmentSuffix(t *testing.T) {
	provider := &staticProvider{
		storage: environments.NewStatic([]string{testStageEnv}, testProdSuffix),
	}

	resolver := NewResolver(nil, nil, provider)

	prod, err := resolver.Runtime(context.Background(), environments.DefaultEnvironmentName)
	require.NoError(t, err)

	stage, err := resolver.Runtime(context.Background(), testStageEnv)
	require.NoError(t, err)

	prodRuntime, ok := prod.(*dockerRuntime)
	require.True(t, ok)

	stageRuntime, ok := stage.(*dockerRuntime)
	require.True(t, ok)

	prodResolver, ok := prodRuntime.resolver.(*labelSuffixResolver)
	require.True(t, ok)

	stageResolver, ok := stageRuntime.resolver.(*labelSuffixResolver)
	require.True(t, ok)

	require.Equal(t, testProdSuffix, prodResolver.suffix)
	require.Equal(t, testStageEnv, stageResolver.suffix)
}

// An empty environment means the default one - the backward-compatibility path
// every pre-environments caller takes.
func TestResolver_EmptyEnvironmentResolvesDefault(t *testing.T) {
	provider := &staticProvider{
		storage: environments.NewStatic(nil, testProdSuffix),
	}

	resolved, err := NewResolver(nil, nil, provider).Runtime(context.Background(), "")
	require.NoError(t, err)

	runtime, ok := resolved.(*dockerRuntime)
	require.True(t, ok)

	resolver, ok := runtime.resolver.(*labelSuffixResolver)
	require.True(t, ok)

	require.Equal(t, testProdSuffix, resolver.suffix)
}

// Best-effort default resolution: with no environments storage at all (boot
// time, single-node before the storage swap) the default degrades to the
// unsuffixed runtime rather than failing.
func TestResolver_NoProviderStillServesDefaultEnvironment(t *testing.T) {
	resolved, err := NewResolver(nil, nil, nil).Runtime(context.Background(), "")
	require.NoError(t, err)

	runtime, ok := resolved.(*dockerRuntime)
	require.True(t, ok)

	resolver, ok := runtime.resolver.(*labelSuffixResolver)
	require.True(t, ok)

	require.Empty(t, resolver.suffix)
}

// An explicitly named environment is strict: unknown never silently falls back
// onto another environment's containers.
func TestResolver_UnknownEnvironmentErrors(t *testing.T) {
	provider := &staticProvider{
		storage: environments.NewStatic(nil, testProdSuffix),
	}

	_, err := NewResolver(nil, nil, provider).Runtime(context.Background(), "ghost")
	require.Error(t, err)
}

// Phase 2's tier: recognised, explicitly refused, never silently served off
// the shared daemon.
func TestResolver_DedicatedEnvironmentIsNotImplemented(t *testing.T) {
	env := domain.Environment{
		ID:         1,
		Name:       testStageEnv,
		Suffix:     testStageEnv,
		DockerHost: "unix:///var/run/docker-stage.sock",
	}

	provider := &staticProvider{
		storage: &dedicatedStorage{env: env},
	}

	_, err := NewResolver(nil, nil, provider).Runtime(context.Background(), testStageEnv)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrDedicatedRuntimeNotImplemented))
}

// The resolver re-reads storage on every call, so an environment's suffix
// change takes effect on the next call without any explicit swap.
func TestResolver_RereadsStorageOnEveryCall(t *testing.T) {
	ctx := context.Background()
	envStorage := environments.NewStatic([]string{testStageEnv}, testProdSuffix)

	provider := &staticProvider{
		storage: envStorage,
	}

	resolver := NewResolver(nil, nil, provider)

	before, err := resolver.Runtime(ctx, testStageEnv)
	require.NoError(t, err)

	env, err := envStorage.GetEnvironmentByName(ctx, testStageEnv)
	require.NoError(t, err)

	newSuffix := "moved"

	updateReq := domain.UpdateEnvironmentReq{
		ID:     env.ID,
		Suffix: &newSuffix,
	}

	_, err = envStorage.UpdateEnvironment(ctx, updateReq)
	require.NoError(t, err)

	after, err := resolver.Runtime(ctx, testStageEnv)
	require.NoError(t, err)

	beforeRuntime, ok := before.(*dockerRuntime)
	require.True(t, ok)

	afterRuntime, ok := after.(*dockerRuntime)
	require.True(t, ok)

	beforeResolver, ok := beforeRuntime.resolver.(*labelSuffixResolver)
	require.True(t, ok)

	afterResolver, ok := afterRuntime.resolver.(*labelSuffixResolver)
	require.True(t, ok)

	require.Equal(t, testStageEnv, beforeResolver.suffix)
	require.Equal(t, newSuffix, afterResolver.suffix)
}
