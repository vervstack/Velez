package transfer_out

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ForEachChunk_CutsReaderIntoChunks(t *testing.T) {
	cases := []struct {
		name      string
		input     []byte
		chunkSize int
		want      []int
	}{
		{"empty reader emits nothing", nil, 4, []int{}},
		{"exact multiple", make([]byte, 8), 4, []int{4, 4}},
		{"short tail", make([]byte, 10), 4, []int{4, 4, 2}},
		{"smaller than one chunk", make([]byte, 3), 4, []int{3}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sizes := make([]int, 0)
			collect := func(chunk []byte) error {
				sizes = append(sizes, len(chunk))

				return nil
			}

			err := forEachChunk(bytes.NewReader(tc.input), tc.chunkSize, collect)
			require.NoError(t, err)
			require.Equal(t, tc.want, sizes)
		})
	}
}

func Test_ForEachChunk_KeepsContentOrder(t *testing.T) {
	input := []byte("0123456789")

	var joined []byte

	collect := func(chunk []byte) error {
		joined = append(joined, chunk...)

		return nil
	}

	err := forEachChunk(bytes.NewReader(input), 3, collect)
	require.NoError(t, err)
	require.Equal(t, input, joined)
}

func Test_ForEachChunk_StopsOnEmitError(t *testing.T) {
	calls := 0
	fail := func([]byte) error {
		calls++

		return errTargetClosedEarly
	}

	err := forEachChunk(bytes.NewReader(make([]byte, 10)), 4, fail)
	require.ErrorIs(t, err, errTargetClosedEarly)
	require.Equal(t, 1, calls)
}
