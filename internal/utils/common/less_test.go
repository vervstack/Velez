package common_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/utils/common"
)

func Test_Less_ReturnsSmallerValue(t *testing.T) {
	cases := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"a less than b", 1, 10, 1},
		{"a greater than b", 10, 1, 1},
		{"a equal to b", 5, 5, 5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := common.Less[int](tc.a, tc.b)
			require.Equal(t, tc.want, got)
		})
	}
}
