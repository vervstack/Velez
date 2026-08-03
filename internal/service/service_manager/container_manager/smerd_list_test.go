package container_manager

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

// fakeRuntimeResolver is a container_runtime.RuntimeResolver that resolves the
// environment for real (environments.Resolve, the same function the
// production resolver uses) and then hands back a canned
// fakeListContainerRuntime, so ListSmerds stays unit-testable without a Docker
// daemon or the real labelBasedRuntime.
//
// Environment-name -> suffix resolution edge cases (unknown environment
// rejected, empty environment falls back to default, best-effort default
// without storage, ...) are no longer ContainerManager's responsibility - they
// live in environments.Resolve and are covered by
// internal/clients/node_clients/container_runtime's own tests
// (resolver_test.go). This fake only proves ListSmerds asks the resolver for
// the right environment and reacts correctly to what comes back.
type fakeRuntimeResolver struct {
	envs storage.EnvironmentsStorage

	resp []container.Summary
	err  error

	gotEnvironment string
}

func (f *fakeRuntimeResolver) Runtime(
	ctx context.Context,
	environment string,
) (container_runtime.ContainerRuntime, error) {
	f.gotEnvironment = environment

	_, err := environments.Resolve(ctx, f.envs, environment)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment")
	}

	return &fakeListContainerRuntime{resp: f.resp, err: f.err}, nil
}

// fakeListContainerRuntime implements only ListContainers - nothing under
// test here calls ContainerCreate.
type fakeListContainerRuntime struct {
	container_runtime.ContainerRuntime

	resp []container.Summary
	err  error
}

func (f *fakeListContainerRuntime) ListContainers(
	_ context.Context, _ *velez_api.ListSmerds_Request,
) ([]container.Summary, error) {
	return f.resp, f.err
}

func newListManager(resolver *fakeRuntimeResolver) *ContainerManager {
	return &ContainerManager{
		runtimes: resolver,
	}
}

// ListSmerds must pass the request's environment through to the resolver
// unchanged - it no longer resolves the suffix itself.
func TestListSmerds_PassesEnvironmentToResolver(t *testing.T) {
	resolver := &fakeRuntimeResolver{envs: environments.NewStatic([]string{"STAGE"}, "prod-suffix")}

	req := &velez_api.ListSmerds_Request{Environment: "STAGE"}

	_, err := newListManager(resolver).ListSmerds(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "STAGE", resolver.gotEnvironment)
}

func TestListSmerds_UnknownEnvironmentRejected(t *testing.T) {
	resolver := &fakeRuntimeResolver{envs: environments.NewStatic(nil, "")}

	req := &velez_api.ListSmerds_Request{Environment: "ghost"}

	_, err := newListManager(resolver).ListSmerds(context.Background(), req)
	require.Error(t, err)
}

func TestListSmerds_FiltersNonVelezContainers(t *testing.T) {
	resolver := &fakeRuntimeResolver{
		envs: environments.NewStatic(nil, "prod-suffix"),
		resp: []container.Summary{
			{ID: "a", Names: []string{"/mine"}, Labels: map[string]string{labels.CreatedWithVelezLabel: labelTrue}},
			{ID: "b", Names: []string{"/foreign"}},
		},
	}

	req := &velez_api.ListSmerds_Request{Environment: environments.DefaultEnvironmentName}

	resp, err := newListManager(resolver).ListSmerds(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.GetSmerds(), 1)
	require.Equal(t, "mine", resp.GetSmerds()[0].GetName())
}

func TestListSmerds_RuntimeErrorPropagates(t *testing.T) {
	resolver := &fakeRuntimeResolver{
		envs: environments.NewStatic(nil, "prod-suffix"),
		err:  rerrors.New("docker down"),
	}

	req := &velez_api.ListSmerds_Request{Environment: environments.DefaultEnvironmentName}

	_, err := newListManager(resolver).ListSmerds(context.Background(), req)
	require.Error(t, err)
}
