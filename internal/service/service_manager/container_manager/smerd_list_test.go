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

// testStageEnvironment/testNetworkName are shared string fixtures for this
// package's tests (smerd_list_test.go, smerds_drop_test.go, network_test.go)
// - factored out to satisfy goconst, which flags the same literal repeated
// across this package's test files.
const (
	testStageEnvironment = "STAGE"
	testNetworkName      = "net1"
	testSmerdName        = "svc"
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

	// inspectResp/inspectFound/inspectErr configure the canned
	// fakeListContainerRuntime's Inspect response for inspect_test.go's
	// InspectSmerd tests - unused (zero value) by ListSmerds's own tests.
	inspectResp  container.InspectResponse
	inspectFound bool
	inspectErr   error

	// resolveErr, when set, short-circuits Runtime() into returning
	// (nil, resolveErr) before doing any of the normal environments.Resolve +
	// fake-construction work below - used by tests that only care about
	// simulating "environment resolution itself failed" (e.g. unknown
	// environment), independent of environments.Resolve's own behavior.
	resolveErr error

	gotEnvironment string

	// removeErrs/connectErr/disconnectErr configure the canned
	// fakeListContainerRuntime's Remove/ConnectToNetwork/
	// DisconnectFromNetworks behavior for smerds_drop_test.go and
	// network_test.go - unused (zero value) by ListSmerds/InspectSmerd's own
	// tests.
	removeErrs map[string]error

	connectErr    error
	disconnectErr error

	// runtime is the most recently constructed fakeListContainerRuntime,
	// exposed so tests can inspect what was recorded on it (removedCalls,
	// gotConnectReq, gotDisconnectContainerID, gotDisconnectNetworks, ...)
	// after the call under test returns.
	runtime *fakeListContainerRuntime
}

func (f *fakeRuntimeResolver) Runtime(
	ctx context.Context,
	environment string,
) (container_runtime.ContainerRuntime, error) {
	f.gotEnvironment = environment

	if f.resolveErr != nil {
		return nil, f.resolveErr
	}

	_, err := environments.Resolve(ctx, f.envs, environment)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment")
	}

	rt := &fakeListContainerRuntime{
		resp: f.resp, err: f.err,
		inspectResp: f.inspectResp, inspectFound: f.inspectFound, inspectErr: f.inspectErr,
		removeErrs:    f.removeErrs,
		connectErr:    f.connectErr,
		disconnectErr: f.disconnectErr,
	}

	f.runtime = rt

	return rt, nil
}

// fakeListContainerRuntime implements ListContainers, Inspect, Remove,
// ConnectToNetwork and DisconnectFromNetworks - the methods exercised by this
// package's tests (ListSmerds/InspectSmerd/DropSmerds/ConnectToNetwork/
// DisconnectFromNetwork). Nothing under test in this package calls
// ContainerCreate.
type fakeListContainerRuntime struct {
	container_runtime.ContainerRuntime

	resp []container.Summary
	err  error

	inspectResp  container.InspectResponse
	inspectFound bool
	inspectErr   error

	// removeErrs is keyed by the identifier passed to Remove; a missing key
	// means Remove succeeds for that identifier.
	removeErrs map[string]error
	// removedCalls records every identifier Remove was called with, in call
	// order.
	removedCalls []string

	connectErr    error
	disconnectErr error

	gotConnectReq container_runtime.ConnectToNetworkRequest

	gotDisconnectContainerID string
	gotDisconnectNetworks    []string
}

func (f *fakeListContainerRuntime) ListContainers(
	_ context.Context, _ *velez_api.ListSmerds_Request,
) ([]container.Summary, error) {
	return f.resp, f.err
}

func (f *fakeListContainerRuntime) Inspect(
	_ context.Context, _ string,
) (container.InspectResponse, bool, error) {
	return f.inspectResp, f.inspectFound, f.inspectErr
}

func (f *fakeListContainerRuntime) Remove(_ context.Context, identifier string) error {
	f.removedCalls = append(f.removedCalls, identifier)

	return f.removeErrs[identifier]
}

func (f *fakeListContainerRuntime) ConnectToNetwork(
	_ context.Context, req container_runtime.ConnectToNetworkRequest,
) error {
	f.gotConnectReq = req

	return f.connectErr
}

func (f *fakeListContainerRuntime) DisconnectFromNetworks(
	_ context.Context, containerID string, networks []string,
) error {
	f.gotDisconnectContainerID = containerID
	f.gotDisconnectNetworks = networks

	return f.disconnectErr
}

func newListManager(resolver *fakeRuntimeResolver) *ContainerManager {
	return &ContainerManager{
		runtimes: resolver,
	}
}

// ListSmerds must pass the request's environment through to the resolver
// unchanged - it no longer resolves the suffix itself.
func TestListSmerds_PassesEnvironmentToResolver(t *testing.T) {
	resolver := &fakeRuntimeResolver{envs: environments.NewStatic([]string{testStageEnvironment}, "prod-suffix")}

	req := &velez_api.ListSmerds_Request{Environment: testStageEnvironment}

	_, err := newListManager(resolver).ListSmerds(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, testStageEnvironment, resolver.gotEnvironment)
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
