package labels

import (
	"strings"
)

// ServiceName is the service a container belongs to: its VervServiceLabel
// value, falling back to its first container name for containers created
// before the label existed.
func ServiceName(containerLabels map[string]string, containerNames []string) string {
	name := containerLabels[VervServiceLabel]
	if name == "" && len(containerNames) != 0 {
		name = strings.TrimPrefix(containerNames[0], "/")
	}

	return name
}

// ResourceTypeOf reports whether a container belongs to serviceName and, when
// its name has the "<serviceName>_" prefix, the resource type (the suffix).
// A container with a VervServiceLabel belongs to exactly the service the label
// names; only an unlabelled container is matched by the name prefix.
func ResourceTypeOf(serviceName string, containerLabels map[string]string, containerName string) (string, bool) {
	resourceType, hasPrefix := strings.CutPrefix(containerName, serviceName+"_")
	if !hasPrefix {
		resourceType = ""
	}

	label := containerLabels[VervServiceLabel]
	if label != "" {
		return resourceType, label == serviceName
	}

	return resourceType, hasPrefix
}
