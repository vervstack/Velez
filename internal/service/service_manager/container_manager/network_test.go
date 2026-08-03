package container_manager

import (
	"context"
	"errors"
	"testing"

	"github.com/docker/docker/errdefs"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/environments"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func TestConnectToNetwork_PassesEnvironmentToResolver(t *testing.T) {
	resolver := &fakeRuntimeResolver{envs: environments.NewStatic([]string{testStageEnvironment}, "prod-suffix")}

	req := domain.Connection{Environment: testStageEnvironment, SmerdName: testSmerdName, Network: testNetworkName}

	err := newListManager(resolver).ConnectToNetwork(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, testStageEnvironment, resolver.gotEnvironment)
}

func TestConnectToNetwork_CallsRuntimeWithMappedRequest(t *testing.T) {
	resolver := &fakeRuntimeResolver{envs: environments.NewStatic(nil, "prod-suffix")}

	req := domain.Connection{SmerdName: testSmerdName, Network: testNetworkName, Aliases: []string{"a1"}}

	err := newListManager(resolver).ConnectToNetwork(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, testSmerdName, resolver.runtime.gotConnectReq.ContainerID)
	require.Equal(t, testNetworkName, resolver.runtime.gotConnectReq.NetworkName)
	require.Equal(t, []string{"a1"}, resolver.runtime.gotConnectReq.Aliases)
}

func TestConnectToNetwork_NetworkNotFoundMapsToErrNetworkNotFound(t *testing.T) {
	resolver := &fakeRuntimeResolver{
		envs:       environments.NewStatic(nil, "prod-suffix"),
		connectErr: errdefs.NotFound(rerrors.New("network x not found")),
	}

	req := domain.Connection{SmerdName: testSmerdName, Network: testNetworkName}

	err := newListManager(resolver).ConnectToNetwork(context.Background(), req)
	require.Error(t, err)
	require.True(t, errors.Is(err, user_errors.ErrNetworkNotFound))
}

func TestConnectToNetwork_UnknownEnvironmentRejected(t *testing.T) {
	resolver := &fakeRuntimeResolver{
		envs:       environments.NewStatic(nil, "prod-suffix"),
		resolveErr: rerrors.New("no such environment"),
	}

	req := domain.Connection{SmerdName: testSmerdName, Network: testNetworkName}

	err := newListManager(resolver).ConnectToNetwork(context.Background(), req)
	require.Error(t, err)
}

func TestDisconnectFromNetwork_PassesEnvironmentToResolver(t *testing.T) {
	resolver := &fakeRuntimeResolver{envs: environments.NewStatic([]string{testStageEnvironment}, "prod-suffix")}

	req := domain.Connection{Environment: testStageEnvironment, SmerdName: testSmerdName, Network: testNetworkName}

	err := newListManager(resolver).DisconnectFromNetwork(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, testStageEnvironment, resolver.gotEnvironment)
	require.Equal(t, testSmerdName, resolver.runtime.gotDisconnectContainerID)
	require.Equal(t, []string{testNetworkName}, resolver.runtime.gotDisconnectNetworks)
}

func TestDisconnectFromNetwork_AlreadyDisconnected_TreatedAsSuccess(t *testing.T) {
	resolver := &fakeRuntimeResolver{
		envs:          environments.NewStatic(nil, "prod-suffix"),
		disconnectErr: rerrors.New("network x is not connected to the network"),
	}

	req := domain.Connection{SmerdName: testSmerdName, Network: testNetworkName}

	err := newListManager(resolver).DisconnectFromNetwork(context.Background(), req)
	require.NoError(t, err)
}

func TestDisconnectFromNetwork_UnexpectedErrorPropagates(t *testing.T) {
	resolver := &fakeRuntimeResolver{
		envs:          environments.NewStatic(nil, "prod-suffix"),
		disconnectErr: rerrors.New("boom"),
	}

	req := domain.Connection{SmerdName: testSmerdName, Network: testNetworkName}

	err := newListManager(resolver).DisconnectFromNetwork(context.Background(), req)
	require.Error(t, err)
}
