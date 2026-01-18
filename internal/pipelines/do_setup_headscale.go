package pipelines

import (
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/patterns/headscale"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/config_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/container_steps"
)

func (p *pipeliner) SetupHeadscale(req domain.SetupHeadscaleRequest) Runner[domain.SetupHeadscaleResponse] {
	contReq := headscale.Headscale(req)
	var containerId string
	mountPoint := &domain.FileMountPoint{
		FilePath: nil,
		Content:  headscale.BasicConfig(),
	}

	return &runner[domain.SetupHeadscaleResponse]{
		Steps: []steps.Step{
			container_steps.Create(p.nodeClients, &contReq, toolbox.ToPtr(headscale.ServiceName),
				&containerId),
			container_steps.CopyToContainer(p.nodeClients, &containerId, mountPoint),
			config_steps.StoreConfig(p.clusterClients,
				headscale.ServiceName, mountPoint.Content,
				matreshka_api.ConfigTypePrefix_plain, matreshka_api.Format_yaml),
		},
	}
}
