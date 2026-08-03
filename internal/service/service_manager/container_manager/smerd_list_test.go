package container_manager

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

// fakeListDocker records the suffix ListSmerds resolved and passed down.
type fakeListDocker struct {
	node_clients.Docker

	gotSuffix string
	resp      []container.Summary
	err       error
}

func (f *fakeListDocker) ListContainers(
	_ context.Context, _ *velez_api.ListSmerds_Request, suffix string,
) ([]container.Summary, error) {
	f.gotSuffix = suffix

	return f.resp, f.err
}

type fakeEnvProvider struct {
	storage storage.EnvironmentsStorage
}

func (f *fakeEnvProvider) Environments() storage.EnvironmentsStorage {
	return f.storage
}

func newListManager(docker node_clients.Docker, envStorage storage.EnvironmentsStorage) *ContainerManager {
	return &ContainerManager{
		dockerWrapper: docker,
		environments:  &fakeEnvProvider{storage: envStorage},
	}
}

// The request's environment NAME must be resolved into the environment's
// SUFFIX before it reaches Docker - the two are not interchangeable.
func TestListSmerds_ResolvesEnvironmentNameToSuffix(t *testing.T) {
	docker := &fakeListDocker{}
	cm := newListManager(docker, environments.NewStatic(nil, "prod-suffix"))

	req := &velez_api.ListSmerds_Request{Environment: environments.DefaultEnvironmentName}

	_, err := cm.ListSmerds(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "prod-suffix", docker.gotSuffix)
}

// Callers that predate environments (the ShutDownOnExit smerd dropper, the UI
// before it learned about environments, e2e tests) pass no environment: that
// must resolve to the DEFAULT environment's suffix - the node's
// pre-environments ContainerSuffix - rather than be rejected or silently
// unscoped.
func TestListSmerds_EmptyEnvironmentUsesDefaultEnvironmentSuffix(t *testing.T) {
	docker := &fakeListDocker{}
	cm := newListManager(docker, environments.NewStatic(nil, "prod-suffix"))

	req := &velez_api.ListSmerds_Request{}

	_, err := cm.ListSmerds(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "prod-suffix", docker.gotSuffix)
}

// On a node that never configured a ContainerSuffix the default environment's
// suffix is empty, so an environment-less list stays unscoped - exactly what it
// did before environments existed.
func TestListSmerds_EmptyEnvironmentOnUnsuffixedNodeStaysUnscoped(t *testing.T) {
	docker := &fakeListDocker{}
	cm := newListManager(docker, environments.NewStatic(nil, ""))

	req := &velez_api.ListSmerds_Request{}

	_, err := cm.ListSmerds(context.Background(), req)
	require.NoError(t, err)
	require.Empty(t, docker.gotSuffix)
}

// Best-effort default resolution: a node whose environments storage isn't up
// yet must still serve environment-less lists instead of erroring.
func TestListSmerds_EmptyEnvironmentWithoutStorageStaysUnscoped(t *testing.T) {
	docker := &fakeListDocker{}
	cm := newListManager(docker, nil)

	req := &velez_api.ListSmerds_Request{}

	_, err := cm.ListSmerds(context.Background(), req)
	require.NoError(t, err)
	require.Empty(t, docker.gotSuffix)
}

// An explicitly named environment must not be confused with the default one:
// each gets its own suffix, which is what keeps their containers apart.
func TestListSmerds_ExplicitEnvironmentGetsOwnSuffix(t *testing.T) {
	docker := &fakeListDocker{}
	cm := newListManager(docker, environments.NewStatic([]string{"STAGE"}, "prod-suffix"))

	req := &velez_api.ListSmerds_Request{Environment: "STAGE"}

	_, err := cm.ListSmerds(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "STAGE", docker.gotSuffix)
}

func TestListSmerds_UnknownEnvironmentRejected(t *testing.T) {
	docker := &fakeListDocker{}
	cm := newListManager(docker, environments.NewStatic(nil, ""))

	req := &velez_api.ListSmerds_Request{Environment: "ghost"}

	_, err := cm.ListSmerds(context.Background(), req)
	require.Error(t, err)
	require.Empty(t, docker.gotSuffix, "Docker must not be queried for an unknown environment")
}

func TestListSmerds_FiltersNonVelezContainers(t *testing.T) {
	docker := &fakeListDocker{
		resp: []container.Summary{
			{ID: "a", Names: []string{"/mine"}, Labels: map[string]string{labels.CreatedWithVelezLabel: labelTrue}},
			{ID: "b", Names: []string{"/foreign"}},
		},
	}
	cm := newListManager(docker, environments.NewStatic(nil, "prod-suffix"))

	req := &velez_api.ListSmerds_Request{Environment: environments.DefaultEnvironmentName}

	resp, err := cm.ListSmerds(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.GetSmerds(), 1)
	require.Equal(t, "mine", resp.GetSmerds()[0].GetName())
}

func TestListSmerds_DockerErrorPropagates(t *testing.T) {
	docker := &fakeListDocker{err: rerrors.New("docker down")}
	cm := newListManager(docker, environments.NewStatic(nil, "prod-suffix"))

	req := &velez_api.ListSmerds_Request{Environment: environments.DefaultEnvironmentName}

	_, err := cm.ListSmerds(context.Background(), req)
	require.Error(t, err)
}
