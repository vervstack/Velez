package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// checkVervLabels asserts the product-stamped labels - the ones Velez itself
// writes (internal/domain/labels), independent of anything the caller passed
// in the request - are present and correct on got. want is built by the
// caller, typically GetExpectedLabels(t) plus overrides (e.g. SuffixLabel for
// a non-default environment); only the keys present in want are checked, so a
// caller only asserts what it actually knows to expect.
func checkVervLabels(t *testing.T, got, want map[string]string) {
	t.Helper()

	for k, v := range want {
		require.Equal(t, v, got[k], "verv-stamped label %q", k)
	}
}

// checkClientLabels asserts labels the caller put in the request (e.g. team,
// verv.tag.canary, or any entry from the request's own Labels field)
// round-tripped onto the container unchanged, independent of the
// verv-stamped labels. Only the keys present in want are checked.
func checkClientLabels(t *testing.T, got, want map[string]string) {
	t.Helper()

	for k, v := range want {
		require.Equal(t, v, got[k], "client label %q", k)
	}
}

// checkPorts asserts got and want describe the same ports, in order. A port's
// ExposedTo is host-assigned at create time, so it's only asserted non-nil
// and >= the DinD-published port band's floor; every other field is compared
// exactly. Mirrors the ports comparison AssertSmerds has always done.
func checkPorts(t *testing.T, got, want []*velez_api.Port) {
	t.Helper()

	require.Len(t, got, len(want))

	for idx, port := range got {
		require.NotNil(t, port.ExposedTo)
		require.GreaterOrEqual(t, port.GetExposedTo(), minPortToExposeTo)

		want[idx].ExposedTo = port.ExposedTo

		if !proto.Equal(want[idx], port) {
			require.Equal(t, want[idx], port)
		}
	}
}

// checkVolumes asserts got and want describe the same volumes, in order.
// Unlike ports, nothing about a Volume is host-assigned - every field is
// compared exactly.
//
// No e2e case creates/mounts a volume yet - tracked as a coverage gap in
// docs/plans/e2e_coverage.md §2.1 "Volumes: create / mount / CopyToVolume".
// Kept ready for whichever suite closes that gap.
//
//nolint:unused // see comment above
func checkVolumes(t *testing.T, got, want []*velez_api.Volume) {
	t.Helper()

	require.Len(t, got, len(want))

	for idx, volume := range got {
		if !proto.Equal(want[idx], volume) {
			require.Equal(t, want[idx], volume)
		}
	}
}
