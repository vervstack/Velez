package transfer_out

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
)

func runningInspect(name string, isRunning bool) container.InspectResponse {
	inspected := newInspect(name, "img:1")

	inspected.State = &container.State{Running: isRunning}

	return inspected
}

func Test_NewSourceOrder_StopsSidecarsFirstRestoresRootFirst(t *testing.T) {
	root := runningInspect("app", true)
	sidecars := []container.InspectResponse{
		runningInspect(agentName, true),
		runningInspect("idle", false),
	}

	order := newSourceOrder(root, sidecars)

	require.Equal(t, []string{agentName, "app"}, order.stop)
	require.Equal(t, []string{"app", agentName}, order.restore)
}

func Test_NewSourceOrder_SkipsStoppedRoot(t *testing.T) {
	order := newSourceOrder(runningInspect("app", false), nil)

	require.Empty(t, order.stop)
	require.Empty(t, order.restore)
}

const (
	agentName = "agent"
)
