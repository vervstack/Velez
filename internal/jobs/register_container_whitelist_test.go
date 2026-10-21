package jobs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

func Test_DecideWhitelistAction_Table(t *testing.T) {
	cases := []struct {
		name              string
		isElevatedAllowed bool
		isAccessRequested bool
		isCluster         bool
		want              whitelistAction
	}{
		{"allowed container needs nothing", true, false, false, whitelistActionNone},
		{"allowed container ignores the flag in cluster mode", true, true, true, whitelistActionNone},
		{"not allowed without the flag is refused in single-node mode", false, false, false, whitelistActionRefuse},
		{"not allowed without the flag is refused in cluster mode", false, false, true, whitelistActionRefuse},
		{"not allowed with the flag adds a row in cluster mode", false, true, true, whitelistActionAddRepository},
		{"not allowed with the flag needs no row in single-node mode", false, true, false, whitelistActionNone},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := decideWhitelistAction(tc.isElevatedAllowed, tc.isAccessRequested, tc.isCluster)

			require.Equal(t, tc.want, got)
		})
	}
}

func Test_StampElevatedAccessAllowance_Table(t *testing.T) {
	cases := []struct {
		name              string
		isAccessRequested bool
		isCluster         bool
		isStamped         bool
	}{
		{"requested in single-node mode stamps the label", true, false, true},
		{"requested in cluster mode leaves the row to the whitelist", true, true, false},
		{"not requested in single-node mode stamps nothing", false, false, false},
		{"not requested in cluster mode stamps nothing", false, true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			extraLabels := map[string]string{}

			stampElevatedAccessAllowance(extraLabels, tc.isAccessRequested, tc.isCluster)

			value, isPresent := extraLabels[labels.ElevatedAccessLabel]
			require.Equal(t, tc.isStamped, isPresent)

			if tc.isStamped {
				require.Equal(t, labels.ElevatedAccessAllowedValue, value)
			}
		})
	}
}
