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

func Test_ToBasicNodeInfo_PassesThroughUsageFieldsUnchanged(t *testing.T) {
	node := domain.NodeBaseInfo{
		Id:            42,
		Name:          "node-42",
		Addr:          "10.0.0.42",
		IsEnabled:     true,
		LastOnline:    time.Now(),
		CpuPercent:    73.5,
		MemPercent:    88.2,
		ServicesCount: 7,
	}

	info := ToBasicNodeInfo(node)

	require.Equal(t, 73.5, info.GetCpuPercent())
	require.Equal(t, 88.2, info.GetMemPercent())
	require.Equal(t, uint64(7), info.GetServicesCount())
}

func Test_ToBasicNodeInfo_PassesThroughRegionUnchanged(t *testing.T) {
	node := domain.NodeBaseInfo{
		Id:         42,
		Name:       "node-42",
		Addr:       "10.0.0.42",
		IsEnabled:  true,
		LastOnline: time.Now(),
		Region:     "eu-west",
	}

	info := ToBasicNodeInfo(node)

	require.Equal(t, "eu-west", info.GetRegion())
}
