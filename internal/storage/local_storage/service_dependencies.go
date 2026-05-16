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
		if dependsOn == "" {
			continue
		}

		targets := strings.Split(dependsOn, ",")
		for _, target := range targets {
			target = strings.TrimSpace(target)
			if target == "" {
				continue
			}

			key := target
			if seen[key] {
				continue
			}
			seen[key] = true

			dep := domain.ServiceDependency{
				SourceService: serviceName,
				TargetService: target,
			}
			result = append(result, dep)
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
