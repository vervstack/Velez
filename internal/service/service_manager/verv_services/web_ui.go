package verv_services

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/rs/zerolog/log"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/service_manager/address_book"
)

const (
	webUiResourceType = "web_ui"
)

func webUiRootService(resource domain.BoundResource, serviceName string) (string, bool) {
	switch resource.ResourceType {
	case webUiResourceType:
		return serviceName, true
	case domain.S3ResourceType:
		instance, _, hasBucket := strings.Cut(resource.Name, "/")
		if !hasBucket || instance == "" {
			return "", false
		}

		return domain.S3ServiceName(instance), true
	default:
		return "", false
	}
}

func (v *VervService) attachWebUis(ctx context.Context, serviceName string, resources []domain.BoundResource) {
	hasWebUi := false

	for _, resource := range resources {
		_, ok := webUiRootService(resource, serviceName)
		if ok {
			hasWebUi = true

			break
		}
	}

	if !hasWebUi {
		return
	}

	for i := range resources {
		rootService, ok := webUiRootService(resources[i], serviceName)
		if !ok {
			continue
		}

		addresses, err := v.addressBook.Addresses(ctx, rootService, domain.WebUiAddressName)
		if err != nil {
			log.Ctx(ctx).Warn().
				Str("service", rootService).
				Err(err).
				Msg("error getting web ui addresses")

			continue
		}

		resources[i].Addresses = addresses

		for _, address := range addresses {
			if address.Scope != velez_api.AddressScope_ADDRESS_SCOPE_DOCKER {
				continue
			}

			resources[i].WebUiHost = address.Host
			resources[i].WebUiPort = address.Port

			break
		}
	}
}

func webUiSidecars(
	containers []container.Summary,
	serviceName string,
	known []domain.ServiceSidecar,
) []domain.ServiceSidecar {
	seen := make(map[string]struct{}, len(known))
	for _, sidecar := range known {
		seen[sidecar.ContainerId] = struct{}{}
	}

	result := known

	for _, cont := range containers {
		target, ok := address_book.WebUiTargetOf(cont.Labels)
		if !ok || target.RootService != serviceName {
			continue
		}

		_, isKnown := seen[cont.ID]
		if isKnown {
			continue
		}

		seen[cont.ID] = struct{}{}
		result = append(result, toServiceSidecar(cont))
	}

	return result
}
