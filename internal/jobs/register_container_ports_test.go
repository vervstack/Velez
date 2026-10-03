package jobs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func hostPort(port uint32) *velez_api.Port {
	return &velez_api.Port{ServicePortNumber: 80, ExposedTo: &port}
}

func Test_FindForeignOccupiedPort(t *testing.T) {
	cases := []struct {
		name      string
		requested []*velez_api.Port
		occupied  []uint32
		own       []*velez_api.Port
		wantPort  uint32
		wantFound bool
	}{
		{name: "free port", requested: []*velez_api.Port{hostPort(8081)}, occupied: []uint32{8080}},
		{
			name:      "own port reused",
			requested: []*velez_api.Port{hostPort(8080)},
			occupied:  []uint32{8080},
			own:       []*velez_api.Port{hostPort(8080)},
		},
		{
			name:      "foreign occupied",
			requested: []*velez_api.Port{hostPort(8081), hostPort(9090)},
			occupied:  []uint32{9090},
			own:       []*velez_api.Port{hostPort(8081)},
			wantPort:  9090,
			wantFound: true,
		},
		{
			name:      "port without exposed to ignored",
			requested: []*velez_api.Port{{ServicePortNumber: 80}},
			occupied:  []uint32{80},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			port, found := findForeignOccupiedPort(tc.requested, tc.occupied, tc.own)

			require.Equal(t, tc.wantFound, found)
			require.Equal(t, tc.wantPort, port)
		})
	}
}
