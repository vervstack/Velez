package container_manager

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

func TestDropSmerds_PassesEnvironmentToResolver(t *testing.T) {
	resolver := &fakeRuntimeResolver{envs: environments.NewStatic([]string{testStageEnvironment}, "prod-suffix")}

	req := &velez_api.DropSmerd_Request{Environment: testStageEnvironment}

	_, err := newListManager(resolver).DropSmerds(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, testStageEnvironment, resolver.gotEnvironment)
}

func TestDropSmerds_RemovesEachIdentifierThroughResolvedRuntime(t *testing.T) {
	resolver := &fakeRuntimeResolver{envs: environments.NewStatic(nil, "prod-suffix")}

	req := &velez_api.DropSmerd_Request{
		Uuids: []string{"a", "b"},
		Name:  []string{"c"},
	}

	resp, err := newListManager(resolver).DropSmerds(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, resolver.runtime.removedCalls)
	require.Equal(t, []string{"a", "b", "c"}, resp.GetSuccessful())
	require.Empty(t, resp.GetFailed())
}

func TestDropSmerds_PerIdentifierRemoveErrorRecordedInFailed_OthersStillProcessed(t *testing.T) {
	resolver := &fakeRuntimeResolver{
		envs:       environments.NewStatic(nil, "prod-suffix"),
		removeErrs: map[string]error{"b": rerrors.New("boom")},
	}

	req := &velez_api.DropSmerd_Request{
		Uuids: []string{"a", "b", "c"},
	}

	resp, err := newListManager(resolver).DropSmerds(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, []string{"a", "c"}, resp.GetSuccessful())
	require.Len(t, resp.GetFailed(), 1)
	require.Equal(t, "b", resp.GetFailed()[0].GetUuid())
	require.Equal(t, "boom", resp.GetFailed()[0].GetCause())
}

func TestDropSmerds_UnknownEnvironmentRejected_AllIdentifiersRecordedFailed(t *testing.T) {
	resolver := &fakeRuntimeResolver{
		envs:       environments.NewStatic(nil, "prod-suffix"),
		resolveErr: rerrors.New("no such environment"),
	}

	req := &velez_api.DropSmerd_Request{
		Uuids: []string{"a", "b"},
	}

	resp, err := newListManager(resolver).DropSmerds(context.Background(), req)
	require.NoError(t, err)
	require.Empty(t, resp.GetSuccessful())
	require.Len(t, resp.GetFailed(), 2)
}
