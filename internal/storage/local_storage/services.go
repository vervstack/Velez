package local_storage

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage"
)

type dockerServices struct {
	docker node_clients.Docker
}

func newServicesStorage(docker node_clients.Docker) *dockerServices {
	return &dockerServices{docker: docker}
}

func (s *dockerServices) GetByName(ctx context.Context, name string) (domain.Service, error) {
	listReq := &pb.ListSmerds_Request{
		Name: &name,
	}

	containers, err := s.docker.ListContainers(ctx, listReq)
	if err != nil {
		return domain.Service{}, rerrors.Wrap(err)
	}

	if len(containers) == 0 {
		return domain.Service{}, storage.ErrNotFound
	}

	c := containers[0]
	svc := domain.Service{
		ServiceBaseInfo: domain.ServiceBaseInfo{
			Name:      name,
			ImageName: c.Image,
			Status:    containerStateToString(c.State),
		},
		Status: containerStateToDeploymentStatus(c.State),
	}
	return svc, nil
}

func containerStateToDeploymentStatus(state string) pb.DeploymentStatus {
	switch state {
	case "running":
		return pb.DeploymentStatus_RUNNING
	case "dead":
		return pb.DeploymentStatus_FAILED
	case "exited":
		return pb.DeploymentStatus_STOPPED
	default:
		return pb.DeploymentStatus_DEPLOYMENT_STATUS_UNKNOWN
	}
}

func containerStateToString(state string) string {
	switch state {
	case "running":
		return "running"
	case "paused":
		return "degraded"
	case "exited", "dead":
		return "stopped"
	default:
		return ""
	}
}

func (s *dockerServices) UpsertService(_ context.Context, _ string) error {
	return nil
}

func (s *dockerServices) List(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error) {
	listReq := &pb.ListSmerds_Request{}

	containers, err := s.docker.ListContainers(ctx, listReq)
	if err != nil {
		return domain.ServiceList{}, rerrors.Wrap(err)
	}

	seen := make(map[string]bool)
	var all []domain.ServiceBaseInfo

	for _, c := range containers {
		serviceName := c.Labels[labels.VervServiceLabel]
		if serviceName == "" || seen[serviceName] {
			continue
		}
		seen[serviceName] = true

		info := domain.ServiceBaseInfo{
			Name: serviceName,
		}
		all = append(all, info)
	}

	if req.NamePattern.Valid {
		pattern := strings.ToLower(req.NamePattern.Value)
		filtered := all[:0]
		for _, svc := range all {
			if strings.Contains(strings.ToLower(svc.Name), pattern) {
				filtered = append(filtered, svc)
			}
		}
		all = filtered
	}

	total := uint64(len(all))

	if req.Paging.Offset < total {
		all = all[req.Paging.Offset:]
	} else {
		all = nil
	}

	if req.Paging.Limit > 0 && uint64(len(all)) > req.Paging.Limit {
		all = all[:req.Paging.Limit]
	}

	out := domain.ServiceList{
		Total:    total,
		Services: all,
	}
	return out, nil
}
