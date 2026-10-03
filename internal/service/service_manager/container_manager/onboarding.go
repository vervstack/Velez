package container_manager

import (
	"github.com/docker/docker/api/types/container"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

// replacedContainerIds maps the id of a replaced container to the id of the container that replaced it.
func replacedContainerIds(list []container.Summary) map[string]string {
	replaced := make(map[string]string)

	for _, cont := range list {
		oldId, isOnboarded := cont.Labels[labels.OnboardedFromLabel]
		if !isOnboarded {
			continue
		}

		replaced[oldId] = cont.ID
	}

	return replaced
}
