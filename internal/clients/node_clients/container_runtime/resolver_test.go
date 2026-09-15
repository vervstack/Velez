package container_runtime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testProdSuffix    = "prod"
	testStageEnv      = "STAGE"
	testDedicatedHost = "tcp://dedicated-stage:2375"
)

type staticProvider struct {
	storage storage.EnvironmentsStorage
}

func (p *staticProvider) Environments() storage.EnvironmentsStorage {
	return p.storage
}

// dedicatedStorage yields a single environment bound to a Docker host other
// than the node's own - only postgres-backed storage can actually persist
// this (static rejects every write), so the resolver's dedicated branch is
// tested against a hand-rolled double rather than environments.NewStatic.
type dedicatedStorage struct {
	storage.EnvironmentsStorage

	env domain.Environment
}

func (s *dedicatedStorage) GetEnvironmentByName(_ context.Context, name string) (domain.Environment, error) {
	if name != s.env.Name {
		return domain.Environment{}, user_errors.ErrStorageNotFound
	}

	return s.env, nil
}

func TestResolver_ResolvesLabelBasedRuntimeWithEnvironmentSuffix(t *testing.T) {
	provider := &staticProvider{
		storage: environments.NewStatic([]string{testStageEnv}, testProdSuffix),
	}

	resolver := NewResolver(nil, "", nil, provider)

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

	resolved, err := NewResolver(nil, "", nil, provider).Runtime(context.Background(), "")
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
	resolved, err := NewResolver(nil, "", nil, nil).Runtime(context.Background(), "")
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

	_, err := NewResolver(nil, "", nil, provider).Runtime(context.Background(), "ghost")
	require.Error(t, err)
}

// An environment whose DockerHost differs from the node's own gets a
// dedicated connection instead of the shared cli - see resolver.dedicatedClient.
func TestResolver_ResolvesDedicatedRuntimeForDifferentDockerHost(t *testing.T) {
	env := domain.Environment{
		Id:         1,
		Name:       testStageEnv,
		Suffix:     testStageEnv,
		DockerHost: testDedicatedHost,
	}

	provider := &staticProvider{
		storage: &dedicatedStorage{env: env},
	}

	resolved, err := NewResolver(nil, "", nil, provider).Runtime(context.Background(), testStageEnv)
	require.NoError(t, err)

	runtime, ok := resolved.(*dockerRuntime)
	require.True(t, ok)

	_, ok = runtime.resolver.(*directResolver)
	require.True(t, ok)
}

// DockerHost equal to the node's own resolved host means "this environment
// shares the node's daemon" - the default every newly created environment
// gets (verv_services.CreateEnvironment) - not "dedicated to a copy of the
// same address".
func TestResolver_TreatsDockerHostEqualToNodeHostAsShared(t *testing.T) {
	env := domain.Environment{
		Id:         1,
		Name:       testStageEnv,
		Suffix:     testStageEnv,
		DockerHost: testDedicatedHost,
	}

	provider := &staticProvider{
		storage: &dedicatedStorage{env: env},
	}

	resolved, err := NewResolver(nil, testDedicatedHost, nil, provider).Runtime(context.Background(), testStageEnv)
	require.NoError(t, err)

	runtime, ok := resolved.(*dockerRuntime)
	require.True(t, ok)

	_, ok = runtime.resolver.(*labelSuffixResolver)
	require.True(t, ok)
}

// A dedicated Docker host is a real connection, unlike the shared cli the
// resolver is handed already built - it must be built once and reused, not
// reconnected on every Runtime call.
func TestResolver_CachesDedicatedClientPerHost(t *testing.T) {
	env := domain.Environment{
		Id:         1,
		Name:       testStageEnv,
		Suffix:     testStageEnv,
		DockerHost: testDedicatedHost,
	}

	provider := &staticProvider{
		storage: &dedicatedStorage{env: env},
	}

	res := NewResolver(nil, "", nil, provider)

	first, err := res.Runtime(context.Background(), testStageEnv)
	require.NoError(t, err)

	second, err := res.Runtime(context.Background(), testStageEnv)
	require.NoError(t, err)

	firstRuntime, ok := first.(*dockerRuntime)
	require.True(t, ok)

	secondRuntime, ok := second.(*dockerRuntime)
	require.True(t, ok)

	require.Same(t, firstRuntime.cli, secondRuntime.cli)
}

// mutableEnvStorage is a hand-rolled, actually-mutable EnvironmentsStorage
// double for TestResolver_RereadsStorageOnEveryCall - environments.NewStatic
// (single-node/dev's production backend) rejects every write, so it can't
// stand in for "storage the resolver observes a live mutation against."
type mutableEnvStorage struct {
	storage.EnvironmentsStorage

	env domain.Environment
}

func (s *mutableEnvStorage) GetEnvironmentByName(_ context.Context, name string) (domain.Environment, error) {
	if name != s.env.Name {
		return domain.Environment{}, user_errors.ErrStorageNotFound
	}

	return s.env, nil
}

func (s *mutableEnvStorage) UpdateEnvironment(
	_ context.Context, req domain.UpdateEnvironmentReq,
) (domain.Environment, error) {
	if req.Suffix != nil {
		s.env.Suffix = *req.Suffix
	}

	return s.env, nil
}

// The resolver re-reads storage on every call, so an environment's suffix
// change takes effect on the next call without any explicit swap.
func TestResolver_RereadsStorageOnEveryCall(t *testing.T) {
	ctx := context.Background()
	envStorage := &mutableEnvStorage{env: domain.Environment{Id: 1, Name: testStageEnv, Suffix: testStageEnv}}

	provider := &staticProvider{
		storage: envStorage,
	}

	resolver := NewResolver(nil, "", nil, provider)

	before, err := resolver.Runtime(ctx, testStageEnv)
	require.NoError(t, err)

	env, err := envStorage.GetEnvironmentByName(ctx, testStageEnv)
	require.NoError(t, err)

	newSuffix := "moved"

	updateReq := domain.UpdateEnvironmentReq{
		Id:     env.Id,
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
