package local_storage

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"
	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
)

type dockerServiceResourcesStorage struct {
	docker node_clients.Docker
}

func newServiceResourcesStorage(docker node_clients.Docker) *dockerServiceResourcesStorage {
	return &dockerServiceResourcesStorage{docker: docker}
}

func (d *dockerServiceResourcesStorage) UpsertResource(_ context.Context, _, _,
	_ string,
) error {
	return nil
}

func (d *dockerServiceResourcesStorage) GetResources(ctx context.Context,
	serviceName string,
) ([]domain.BoundResource, error) {
	listReq := &pb.ListSmerds_Request{}

	containers, err := d.docker.ListContainers(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	var result []domain.BoundResource

	prefix := serviceName + "_"

	for _, c := range containers {
		if len(c.Names) == 0 {
			continue
		}

		name := strings.TrimPrefix(c.Names[0], "/")
		if !strings.HasPrefix(name, prefix) {
			continue
		}

		resourceType := strings.TrimPrefix(name, prefix)
		status := c.State

		resource := domain.BoundResource{
			Name:         name,
			ResourceType: resourceType,
			Status:       status,
		}

		result = append(result, resource)
	}

	return result, nil
}
