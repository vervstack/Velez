package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func Test_NodeStatusFromLastOnline_DisabledNodeIsOffline(t *testing.T) {
	node := domain.NodeBaseInfo{
		IsEnabled:  false,
		LastOnline: time.Now(),
	}

	status := NodeStatusFromLastOnline(node)

	require.Equal(t, pb.NodeStatus_NodeStatus_Offline, status)
}

func Test_NodeStatusFromLastOnline_FreshHeartbeatIsOnline(t *testing.T) {
	node := domain.NodeBaseInfo{
		IsEnabled:  true,
		LastOnline: time.Now(),
	}

	status := NodeStatusFromLastOnline(node)

	require.Equal(t, pb.NodeStatus_NodeStatus_Online, status)
}

func Test_NodeStatusFromLastOnline_StaleUnderOfflineThresholdIsDegraded(t *testing.T) {
	node := domain.NodeBaseInfo{
		IsEnabled:  true,
		LastOnline: time.Now().Add(-2 * time.Minute),
	}

	status := NodeStatusFromLastOnline(node)

	require.Equal(t, pb.NodeStatus_NodeStatus_Degraded, status)
}

func Test_NodeStatusFromLastOnline_PastOfflineThresholdIsOffline(t *testing.T) {
	node := domain.NodeBaseInfo{
		IsEnabled:  true,
		LastOnline: time.Now().Add(-6 * time.Minute),
	}

	status := NodeStatusFromLastOnline(node)

	require.Equal(t, pb.NodeStatus_NodeStatus_Offline, status)
}
