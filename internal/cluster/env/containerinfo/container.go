// Package containerinfo detects whether the current process is running
// inside a container. It has no internal dependencies beyond toolbox so it
// can be imported from low-level packages (e.g. node_clients/hardware)
// without pulling in package env's node_clients dependency and causing an
// import cycle.
package containerinfo

import (
	"os"
	"strings"

	"go.redsock.ru/toolbox"
)

var instanceContainerID *string

// IsInContainer - function to determine weather
// this instance ran inside a container or as a standalone app
// returns container uuid if so.
func IsInContainer() bool {
	return GetContainerId() != nil
}

func GetContainerId() *string {
	if instanceContainerID == nil {
		instanceContainerID = getContainerID()
		if instanceContainerID == nil {
			instanceContainerID = toolbox.ToPtr("")
		}
	}

	if *instanceContainerID == "" {
		return nil
	}

	return instanceContainerID
}

func getContainerID() *string {
	hm, err := os.ReadFile("/etc/hostname")
	if err != nil {
		return nil
	}

	return toolbox.ToPtr(strings.TrimRight(string(hm), "\n"))
}
