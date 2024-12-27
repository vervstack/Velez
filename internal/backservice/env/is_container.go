package env

import (
	"os"
	"strings"

	"go.redsock.ru/toolbox"
)

var instanceContainerId *string

// IsInContainer - function to determine weather
// this instance ran inside a container or as a standalone app
// returns container uuid if so
func IsInContainer() bool {
	return GetContainerId() != nil
}

func GetContainerId() *string {
	if instanceContainerId == nil {
		instanceContainerId = getContainerId()
		if instanceContainerId == nil {
			instanceContainerId = toolbox.ToPtr("")
		}
	}

	if *instanceContainerId == "" {
		return nil
	}

	return instanceContainerId
}

func getContainerId() *string {
	hm, err := os.ReadFile("/etc/hostname")
	if err != nil {
		return nil
	}
	return toolbox.ToPtr(strings.TrimRight(string(hm), "\n"))
}
