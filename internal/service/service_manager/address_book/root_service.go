package address_book

import (
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

// RootServiceName is the service a container's addresses are registered under:
// a web ui container (legacy or sidecar) belongs to the service it serves,
// any other container to its own verv service.
func RootServiceName(containerLabels map[string]string) string {
	rootService := containerLabels[labels.WebUiForLabel]
	if rootService != "" {
		return rootService
	}

	instance := containerLabels[labels.S3WebUiLabel]
	if instance != "" {
		return domain.S3ServiceName(instance)
	}

	return containerLabels[labels.VervServiceLabel]
}
