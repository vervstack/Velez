package velez_api_impl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortsToRanges_CompressesPools(t *testing.T) {
	cases := []struct {
		name  string
		ports []int
		want  []string
	}{
		{
			name:  "consecutive ports",
			ports: []int{18503, 5432, 18501, 18502},
			want:  []string{"5432", "18501:18503"},
		},
		{
			name:  "empty pool",
			ports: nil,
			want:  []string{},
		},
		{
			name:  "single port",
			ports: []int{5432},
			want:  []string{"5432"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ranges := portsToRanges(tc.ports)

			require.Equal(t, tc.want, ranges)
		})
	}
}
