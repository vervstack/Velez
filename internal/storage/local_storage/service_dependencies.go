package local_storage

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

func containerName(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return strings.TrimPrefix(names[0], "/")
}

type dockerServiceDepsStorage struct {
	docker node_clients.Docker
}

func newServiceDepsStorage(docker node_clients.Docker) *dockerServiceDepsStorage {
	return &dockerServiceDepsStorage{docker: docker}
}

func (d *dockerServiceDepsStorage) UpsertDependency(ctx context.Context, source, target, proto string) error {
	return nil
}

func (d *dockerServiceDepsStorage) GetDependencies(ctx context.Context, serviceName string) ([]domain.ServiceDependency, error) {
	labelFilter := map[string]string{
		labels.VervServiceLabel: serviceName,
	}
	listReq := &pb.ListSmerds_Request{
		Label: labelFilter,
	}

	containers, err := d.docker.ListContainers(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	seen := make(map[string]bool)
	var result []domain.ServiceDependency

	for _, c := range containers {
		dependsOn := c.Labels[labels.DependsOnLabel]
		if dependsOn != "" {
			targets := strings.Split(dependsOn, ",")
			for _, target := range targets {
				target = strings.TrimSpace(target)
				if target == "" {
					continue
				}

				if seen[target] {
					continue
				}
				seen[target] = true

				dep := domain.ServiceDependency{
					SourceService: serviceName,
					TargetService: target,
					NodeType:      domain.NodeTypeService,
				}
				result = append(result, dep)
			}
		} else {
			name := containerName(c.Names)
			if name != "" && name != serviceName {
				dep := domain.ServiceDependency{
					SourceService: serviceName,
					TargetService: name,
					NodeType:      domain.NodeTypeResource,
				}
				result = append(result, dep)
			}
		}
	}

	return result, nil
}

func (d *dockerServiceDepsStorage) GetCallers(ctx context.Context, serviceName string) ([]domain.ServiceDependency, error) {
	listReq := &pb.ListSmerds_Request{}
	containers, err := d.docker.ListContainers(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	var result []domain.ServiceDependency

	for _, c := range containers {
		dependsOn := c.Labels[labels.DependsOnLabel]
		if dependsOn == "" {
			continue
		}

		targets := strings.Split(dependsOn, ",")
		found := false
		for _, target := range targets {
			target = strings.TrimSpace(target)
			if target == serviceName {
				found = true
				break
			}
		}

		if !found {
			continue
		}

		sourceService := c.Labels[labels.VervServiceLabel]
		if sourceService == "" {
			continue
		}

		dep := domain.ServiceDependency{
			SourceService: sourceService,
			TargetService: serviceName,
		}
		result = append(result, dep)
	}

	return result, nil
}
