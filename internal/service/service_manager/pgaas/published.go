package pgaas

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	tcpProtocol = "tcp"
)

type publishedContainer struct {
	environment domain.Environment
	ports       []container.Port
}

func (p publishedContainer) exposedPort(containerPort int32) *uint32 {
	for _, port := range p.ports {
		if int32(port.PrivatePort) == containerPort && port.Type == tcpProtocol && port.PublicPort != 0 {
			exposed := uint32(port.PublicPort)

			return &exposed
		}
	}

	return nil
}

// publishedContainers maps the VervServiceLabel value to the container's port
// bindings, with one container listing per environment.
func (s *PgaasService) publishedContainers(
	ctx context.Context, labelFilter map[string]string,
) (map[string]publishedContainer, error) {
	environments, err := s.dataStorage.Environments().ListEnvironments(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing environments")
	}

	if len(environments) == 0 {
		environments = []domain.Environment{{}}
	}

	out := make(map[string]publishedContainer)

	for _, environment := range environments {
		runtime, runtimeErr := s.runtimes.Runtime(ctx, environment.Name)
		if runtimeErr != nil {
			return nil, rerrors.Wrap(runtimeErr, "error resolving container runtime")
		}

		listReq := &velez_api.ListSmerds_Request{
			Environment: environment.Name,
			Label:       labelFilter,
		}

		list, listErr := runtime.ListContainers(ctx, listReq)
		if listErr != nil {
			return nil, rerrors.Wrap(listErr, "error listing containers")
		}

		for _, summary := range list {
			name, ok := summary.Labels[labels.VervServiceLabel]
			if !ok {
				continue
			}

			out[name] = publishedContainer{environment: environment, ports: summary.Ports}
		}
	}

	return out, nil
}
