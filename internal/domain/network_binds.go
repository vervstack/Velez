package domain

import (
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// MergeNetworkBinds appends every bind of extra whose network name is not
// already in base.
func MergeNetworkBinds(base, extra []*velez_api.NetworkBind) []*velez_api.NetworkBind {
	known := make(map[string]struct{}, len(base))
	for _, bind := range base {
		known[bind.GetNetworkName()] = struct{}{}
	}

	for _, bind := range extra {
		_, isKnown := known[bind.GetNetworkName()]
		if isKnown {
			continue
		}

		base = append(base, bind)
	}

	return base
}
