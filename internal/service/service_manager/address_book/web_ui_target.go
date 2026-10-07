package address_book

import (
	"strconv"

	"github.com/docker/docker/api/types/container"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	tcpProtocol = "tcp"
)

type WebUiTarget struct {
	RootService   string
	ContainerPort uint16
}

func WebUiTargetOf(containerLabels map[string]string) (WebUiTarget, bool) {
	rootService := containerLabels[labels.WebUiForLabel]
	if rootService != "" {
		port, err := strconv.ParseUint(containerLabels[labels.WebUiPortLabel], 10, 16)
		if err != nil {
			return WebUiTarget{}, false
		}

		target := WebUiTarget{
			RootService:   rootService,
			ContainerPort: uint16(port),
		}

		return target, true
	}

	instance := containerLabels[labels.S3WebUiLabel]
	if instance == "" {
		return WebUiTarget{}, false
	}

	target := WebUiTarget{
		RootService:   domain.S3ServiceName(instance),
		ContainerPort: domain.S3WebUiContainerPort,
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

func serviceRoots(containers []container.Summary) map[string]container.Summary {
	roots := make(map[string]container.Summary)

	for _, cont := range containers {
		_, isSidecar := cont.Labels[labels.Sidecar]
		serviceName := cont.Labels[labels.VervServiceLabel]

		if isSidecar || serviceName == "" {
			continue
		}

		roots[serviceName] = cont
	}

	return roots
}
