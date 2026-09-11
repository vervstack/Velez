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

// dockerEnvPath and hostnamePath are vars, not consts, so tests can point
// them at a temp dir instead of the real root filesystem.
var (
	dockerEnvPath       = "/.dockerenv"
	hostnamePath        = "/etc/hostname"
	instanceContainerID *string
)

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

// getContainerID trusts /etc/hostname as the container id only once
// dockerEnvPath confirms we're actually inside a Docker container.
// /etc/hostname exists on every Linux host, containerized or not, so
// checking it alone reports "in container" on any bare Linux host or CI
// runner.
func getContainerID() *string {
	_, err := os.Stat(dockerEnvPath)
	if err != nil {
		return nil
	}

	hm, err := os.ReadFile(hostnamePath)
	if err != nil {
		return nil
	}

	return toolbox.ToPtr(strings.TrimRight(string(hm), "\n"))
}
