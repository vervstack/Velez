//go:build e2e_full

package e2e

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

const (
	containerLogsSuffix = "e2elogs"

	containerLogsNumberedScript = `i=1; while [ $i -le 5 ]; do echo line-$i; i=$((i+1)); done; sleep 3600`
	containerLogsStderrScript   = `echo out-line; echo err-line >&2; sleep 3600`
)

type ContainerLogsSuite struct {
	suite.Suite

	plane Plane
}

func newContainerLogsRequest(containerId string, tail uint32, isFollow bool) *velez_api.WatchContainerLogs_Request {
	return &velez_api.WatchContainerLogs_Request{
		Id:       containerId,
		Tail:     tail,
		IsFollow: isFollow,
	}
}

func startContainerLogsFixture(t *testing.T, env *TestEnvironment, script string) string {
	t.Helper()

	ensureImagePresent(t, dockerOf(env), transferImage)

	cfg := &container.Config{
		Image: transferImage,
		Cmd:   []string{"sh", "-c", script},
	}

	hostCfg := &container.HostConfig{}

	started := startFixtureContainer(t, env, GetServiceName(t), cfg, hostCfg)

	return started.ID
}

func receiveContainerLogs(
	t *testing.T,
	stream grpc.ServerStreamingClient[velez_api.WatchContainerLogs_Response],
	count int,
) []*velez_api.WatchContainerLogs_Response {
	t.Helper()

	received := make([]*velez_api.WatchContainerLogs_Response, 0, count)

	for range count {
		msg, err := stream.Recv()
		require.NoError(t, err)

		received = append(received, msg)
	}

	return received
}

func receiveContainerLogsUntilEof(
	t *testing.T,
	stream grpc.ServerStreamingClient[velez_api.WatchContainerLogs_Response],
) []*velez_api.WatchContainerLogs_Response {
	t.Helper()

	var received []*velez_api.WatchContainerLogs_Response

	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return received
		}

		require.NoError(t, err)

		received = append(received, msg)
	}
}

func containerLogTexts(logs []*velez_api.WatchContainerLogs_Response) []string {
	texts := make([]string, 0, len(logs))

	for _, entry := range logs {
		texts = append(texts, entry.GetText())
	}

	return texts
}

func numberedLines(from, to int) []string {
	lines := make([]string, 0, to-from+1)

	for i := from; i <= to; i++ {
		lines = append(lines, fmt.Sprintf("line-%d", i))
	}

	return lines
}

func (s *ContainerLogsSuite) Test_WatchContainerLogs_FollowStreamsLinesInOrder() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t, WithContainerSuffix(containerLogsSuffix))
	containerId := startContainerLogsFixture(t, env, containerLogsNumberedScript)

	streamCtx, cancel := context.WithCancel(t.Context())
	defer cancel()

	req := newContainerLogsRequest(containerId, 0, true)

	stream, err := velez_api.NewVelezAPIClient(env.grpcConn).WatchContainerLogs(streamCtx, req)
	require.NoError(t, err)

	received := receiveContainerLogs(t, stream, 5)

	require.Equal(t, numberedLines(1, 5), containerLogTexts(received))
	require.NotNil(t, received[0].GetTimestamp())
	require.Equal(t, velez_api.LogStream_STDOUT, received[0].GetStream())
}

func (s *ContainerLogsSuite) Test_WatchContainerLogs_WithoutFollowClosesAfterBacklog() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t, WithContainerSuffix(containerLogsSuffix))
	containerId := startContainerLogsFixture(t, env, containerLogsNumberedScript)

	req := newContainerLogsRequest(containerId, 0, false)

	stream, err := velez_api.NewVelezAPIClient(env.grpcConn).WatchContainerLogs(t.Context(), req)
	require.NoError(t, err)

	received := receiveContainerLogsUntilEof(t, stream)

	require.Equal(t, numberedLines(1, 5), containerLogTexts(received))
}

func (s *ContainerLogsSuite) Test_WatchContainerLogs_TailLimitsBacklog() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t, WithContainerSuffix(containerLogsSuffix))
	containerId := startContainerLogsFixture(t, env, containerLogsNumberedScript)

	req := newContainerLogsRequest(containerId, 2, false)

	stream, err := velez_api.NewVelezAPIClient(env.grpcConn).WatchContainerLogs(t.Context(), req)
	require.NoError(t, err)

	received := receiveContainerLogsUntilEof(t, stream)

	require.Equal(t, numberedLines(4, 5), containerLogTexts(received))
}

func (s *ContainerLogsSuite) Test_WatchContainerLogs_UnknownIdIsNotFound() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t, WithContainerSuffix(containerLogsSuffix))

	req := newContainerLogsRequest("e2e-no-such-container", 0, false)

	stream, err := velez_api.NewVelezAPIClient(env.grpcConn).WatchContainerLogs(t.Context(), req)
	require.NoError(t, err)

	_, err = stream.Recv()
	require.Error(t, err)
	require.Equal(t, codes.NotFound, status.Code(err))
}

func (s *ContainerLogsSuite) Test_WatchContainerLogs_StderrLineIsTaggedStderr() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t, WithContainerSuffix(containerLogsSuffix))
	containerId := startContainerLogsFixture(t, env, containerLogsStderrScript)

	req := newContainerLogsRequest(containerId, 0, false)

	stream, err := velez_api.NewVelezAPIClient(env.grpcConn).WatchContainerLogs(t.Context(), req)
	require.NoError(t, err)

	received := receiveContainerLogsUntilEof(t, stream)

	streamByText := map[string]velez_api.LogStream{}
	for _, entry := range received {
		streamByText[entry.GetText()] = entry.GetStream()
	}

	require.Equal(t, velez_api.LogStream_STDOUT, streamByText["out-line"])
	require.Equal(t, velez_api.LogStream_STDERR, streamByText["err-line"])
}

func Test_ContainerLogs(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &ContainerLogsSuite{plane: plane}
	})
}
