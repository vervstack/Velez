package verv_services

import (
	"context"
	"maps"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	tcpProtocol = "tcp"
)

type webUiTarget struct {
	rootService   string
	containerPort uint16
}

type webUiAddress struct {
	host string
	port uint32
}

func webUiRootService(resource domain.BoundResource) (string, bool) {
	switch resource.ResourceType {
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

func webUiTargetOf(containerLabels map[string]string) (webUiTarget, bool) {
	rootService := containerLabels[labels.WebUiForLabel]
	if rootService != "" {
		port, err := strconv.ParseUint(containerLabels[labels.WebUiPortLabel], 10, 16)
		if err != nil {
			return webUiTarget{}, false
		}

		target := webUiTarget{
			rootService:   rootService,
			containerPort: uint16(port),
		}

		return target, true
	}

	instance := containerLabels[labels.S3WebUiLabel]
	if instance == "" {
		return webUiTarget{}, false
	}

	target := webUiTarget{
		rootService:   domain.S3ServiceName(instance),
		containerPort: domain.S3WebUiContainerPort,
	}

	return target, true
}

func publishedPort(summary container.Summary, private uint16) uint32 {
	for _, port := range summary.Ports {
		if port.PrivatePort == private && port.Type == tcpProtocol && port.PublicPort != 0 {
			return uint32(port.PublicPort)
		}
	}

	return 0
}

func (v *VervService) discoverWebUis(ctx context.Context) (map[string]webUiAddress, error) {
	environments, err := v.dataStorage.Environments().ListEnvironments(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing environments")
	}

	addresses := make(map[string]webUiAddress)

	for _, env := range environments {
		found, discoverErr := v.discoverEnvironmentWebUis(ctx, env)
		if discoverErr != nil {
			log.Ctx(ctx).Warn().
				Str("environment", env.Name).
				Err(discoverErr).
				Msg("error discovering environment web uis")

			continue
		}

		maps.Copy(addresses, found)
	}

	return addresses, nil
}

func (v *VervService) discoverEnvironmentWebUis(
	ctx context.Context,
	env domain.Environment,
) (map[string]webUiAddress, error) {
	runtime, err := v.runtimes.Runtime(ctx, env.Name)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment runtime")
	}

	listReq := &velez_api.ListSmerds_Request{Environment: env.Name}

	containers, err := runtime.ListContainers(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing environment containers")
	}

	host := env.RemoteHost(v.docker.Host())
	addresses := make(map[string]webUiAddress)

	for _, cont := range containers {
		target, ok := webUiTargetOf(cont.Labels)
		if !ok {
			continue
		}

		port := publishedPort(cont, target.containerPort)
		if port == 0 {
			continue
		}

		addresses[target.rootService] = webUiAddress{host: host, port: port}
	}

	return addresses, nil
}

func (v *VervService) attachWebUis(ctx context.Context, resources []domain.BoundResource) {
	hasWebUi := false

	for _, resource := range resources {
		_, ok := webUiRootService(resource)
		if ok {
			hasWebUi = true

			break
		}
	}

	if !hasWebUi {
		return
	}

	addresses, err := v.discoverWebUis(ctx)
	if err != nil {
		log.Ctx(ctx).Warn().
			Err(err).
			Msg("error discovering web uis")

		return
	}

	for i := range resources {
		rootService, ok := webUiRootService(resources[i])
		if !ok {
			continue
		}

		address, found := addresses[rootService]
		if !found {
			continue
		}

		resources[i].WebUiHost = address.host
		resources[i].WebUiPort = address.port
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
		target, ok := webUiTargetOf(cont.Labels)
		if !ok || target.rootService != serviceName {
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
