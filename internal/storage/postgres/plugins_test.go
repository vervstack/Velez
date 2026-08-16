package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func Test_convertNodeIds(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		nodeIds  []int32
		expected []int64
	}{
		{
			name:     "empty input returns nil",
			nodeIds:  nil,
			expected: nil,
		},
		{
			name:     "single node id is converted",
			nodeIds:  []int32{1},
			expected: []int64{1},
		},
		{
			name:     "multiple node ids preserve order",
			nodeIds:  []int32{3, 1, 2},
			expected: []int64{3, 1, 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := convertNodeIds(tc.nodeIds)
			require.Equal(t, tc.expected, got)
		})
	}
}

func Test_calculateMasterNodeId(t *testing.T) {
	t.Parallel()

	statefullPgType := pb.VervPluginType_name[int32(pb.VervPluginType_statefull_pg)]

	testCases := []struct {
		name       string
		pluginType string
		nodeIds    []int64
		expected   *int64
	}{
		{
			name:       "statefull_pg with a running node returns that node as master",
			pluginType: statefullPgType,
			nodeIds:    []int64{42},
			expected:   int64Ptr(42),
		},
		{
			name:       "statefull_pg with no running node returns nil",
			pluginType: statefullPgType,
			nodeIds:    nil,
			expected:   nil,
		},
		{
			name:       "non statefull_pg plugin never gets a master node id",
			pluginType: pb.VervPluginType_name[int32(pb.VervPluginType_headscale)],
			nodeIds:    []int64{42},
			expected:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := calculateMasterNodeId(tc.pluginType, tc.nodeIds)
			require.Equal(t, tc.expected, got)
		})
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
