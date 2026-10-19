package autoupdate

import (
	"context"
	"sort"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

func (s *AutoUpdateServiceImpl) ListServices(ctx context.Context) (*velez_api.ListAutoUpdateServices_Response, error) {
	conts, err := s.listAppContainers(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing app services")
	}

	services := make([]*velez_api.ListAutoUpdateServices_Response_Service, 0, len(conts))

	for _, cont := range conts {
		item := &velez_api.ListAutoUpdateServices_Response_Service{
			Name:      containerName(cont),
			Image:     cont.Image,
			IsEnabled: cont.Labels[labels.AutoUpgrade] == labelTrue,
			BumpLevel: domain.BumpLevelFromLabels(cont.Labels),
		}

		services = append(services, item)
	}

	sort.Slice(services, func(i, j int) bool {
		return services[i].GetName() < services[j].GetName()
	})

	resp := &velez_api.ListAutoUpdateServices_Response{
		Services:           services,
		CheckPeriodSeconds: s.checkPeriodSeconds,
		IsWorkerEnabled:    s.isWorkerEnabled,
	}

	return resp, nil
}
