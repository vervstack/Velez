package env

import (
	"go.vervstack.ru/Velez/internal/cluster/env/containerinfo"
)

// IsInContainer - function to determine whether
// this instance ran inside a container or as a standalone app
// returns container uuid if so.
func IsInContainer() bool {
	return containerinfo.IsInContainer()
}

func GetContainerId() *string {
	return containerinfo.GetContainerId()
}
