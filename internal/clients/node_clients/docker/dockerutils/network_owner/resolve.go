package network_owner

import (
	"strings"

	"github.com/docker/docker/api/types/container"
)

const (
	containerModePrefix = "container:"
)

// OwnerIds maps container id -> id of the root container whose network namespace it shares.
// A container that shares no namespace is absent from the map. Chains (A shares B, B shares C) resolve to C.
func OwnerIds(list []container.Summary) map[string]string {
	direct := make(map[string]string)

	for _, cont := range list {
		target, isShared := strings.CutPrefix(cont.HostConfig.NetworkMode, containerModePrefix)
		if !isShared {
			continue
		}

		ownerId, isFound := findContainerId(list, target)
		if !isFound {
			continue
		}

		direct[cont.ID] = ownerId
	}

	owners := make(map[string]string, len(direct))
	for id := range direct {
		rootId, isResolved := resolveRoot(direct, id)
		if !isResolved {
			continue
		}

		owners[id] = rootId
	}

	return owners
}

// SidecarsOf returns the containers whose resolved root is rootId (rootId itself excluded), in list order.
func SidecarsOf(list []container.Summary, rootId string) []container.Summary {
	owners := OwnerIds(list)

	sidecars := make([]container.Summary, 0)

	for _, cont := range list {
		ownerId, isShared := owners[cont.ID]
		if isShared && ownerId == rootId && cont.ID != rootId {
			sidecars = append(sidecars, cont)
		}
	}

	return sidecars
}

func resolveRoot(direct map[string]string, id string) (string, bool) {
	visited := map[string]struct{}{id: {}}

	current := id
	for {
		next, isShared := direct[current]
		if !isShared {
			return current, true
		}

		_, isCycle := visited[next]
		if isCycle {
			return "", false
		}

		visited[next] = struct{}{}
		current = next
	}
}

func findContainerId(list []container.Summary, target string) (string, bool) {
	if target == "" {
		return "", false
	}

	for _, cont := range list {
		if cont.ID == target {
			return cont.ID, true
		}
	}

	for _, cont := range list {
		for _, name := range cont.Names {
			if strings.TrimPrefix(name, "/") == target {
				return cont.ID, true
			}
		}
	}

	for _, cont := range list {
		if strings.HasPrefix(cont.ID, target) {
			return cont.ID, true
		}
	}

	return "", false
}
