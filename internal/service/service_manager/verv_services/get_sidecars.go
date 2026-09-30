package verv_services

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/network_owner"
	"go.vervstack.ru/Velez/internal/domain"
)

const (
	listAllContainersNoLimit = 0
)

func (v *VervService) listSidecars(ctx context.Context, serviceName string) ([]domain.ServiceSidecar, error) {
	listReq := &velez_api.ListSmerds_Request{
		Name: &serviceName,
	}

	resp, err := v.containerService.ListSmerds(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing smerds for sidecars")
	}

	if len(resp.GetSmerds()) == 0 {
		return nil, nil
	}

	runtime, err := v.runtimes.Runtime(ctx, listReq.GetEnvironment())
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment")
	}

	containers, err := runtime.ListAllContainers(ctx, listAllContainersNoLimit)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	var sidecars []domain.ServiceSidecar

	for _, smerd := range resp.GetSmerds() {
		for _, sidecar := range network_owner.SidecarsOf(containers, smerd.GetUuid()) {
			sidecars = append(sidecars, toServiceSidecar(sidecar))
		}
	}

	return sidecars, nil
}

func toServiceSidecar(summary container.Summary) domain.ServiceSidecar {
	sidecar := domain.ServiceSidecar{
		ContainerId: summary.ID,
		ImageName:   summary.Image,
		Status:      velez_api.Smerd_Status(velez_api.Smerd_Status_value[summary.State]),
	}

	if len(summary.Names) != 0 {
		sidecar.ContainerName = strings.TrimPrefix(summary.Names[0], "/")
	}

	return sidecar
}
