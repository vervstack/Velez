package pipelines

import (
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/service_steps"
)

func (p *pipeliner) CreateService(req domain.CreateServiceReq) Runner[domain.ServiceBaseInfo] {
	serviceInfo := domain.ServiceBaseInfo{
		Name: req.Name,
	}

	var serviceId uint64

	return &runner[domain.ServiceBaseInfo]{
		Steps: []steps.Step{
			service_steps.ValidateServiceName(req.Name),
			service_steps.UpsertServiceState(p.clusterClients.StateManager(), &serviceInfo, &serviceId),
		},
	}
}
