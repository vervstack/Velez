package verv_services

import (
	"context"
	"sort"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

func (v *VervService) GetServiceEnvironments(ctx context.Context, serviceName string) ([]domain.ServiceEnvironment, error) {
	req := &velez_api.ListSmerds_Request{
		Label: map[string]string{labels.VervServiceLabel: serviceName},
	}
	resp, err := v.containerService.ListSmerds(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	envGroups := make(map[string][]*velez_api.Smerd)
	for _, smerd := range resp.Smerds {
		env := smerd.Labels["env"]
		envGroups[env] = append(envGroups[env], smerd)
	}

	result := make([]domain.ServiceEnvironment, 0, len(envGroups))

	for env, smerds := range envGroups {
		if len(smerds) == 0 {
			continue
		}

		runningCount := 0
		for _, smerd := range smerds {
			if smerd.Status == velez_api.Smerd_running {
				runningCount++
			}
		}

		status := "stopped"
		if runningCount == len(smerds) {
			status = "running"
		} else if runningCount > 0 {
			status = "degraded"
		}

		firstSmerd := smerds[0]
		var deployedAt *time.Time
		if firstSmerd.CreatedAt != nil {
			t := firstSmerd.CreatedAt.AsTime()
			deployedAt = &t
		}

		envEntry := domain.ServiceEnvironment{
			Env:             env,
			Status:          status,
			DeployedVersion: firstSmerd.ImageName,
			Health:          status,
			DeployedAt:      deployedAt,
		}
		result = append(result, envEntry)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Env < result[j].Env
	})

	return result, nil
}
