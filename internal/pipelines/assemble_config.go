package pipelines

import (
	"github.com/docker/docker/api/types"
	"go.vervstack.ru/matreshka"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/internal/pipelines/steps"
	"github.com/godverv/Velez/pkg/velez_api"
)

const configFetchingPostfix = "_config_scanning"

func (p *pipeliner) AssembleConfig(req *velez_api.AssembleConfig_Request) Runner[matreshka.AppConfig] {
	image := &types.ImageInspect{}
	contId := ""

	createReq := domain.LaunchSmerd{
		CreateSmerd_Request: &velez_api.CreateSmerd_Request{
			Name:      req.ServiceName + configFetchingPostfix,
			ImageName: req.ImageName,
			Settings:  &velez_api.Container_Settings{},
		}}

	res := &matreshka.AppConfig{}

	return &runner[matreshka.AppConfig]{
		Steps: []steps.Step{
			steps.PrepareImageStep(p.nodeClients, req.ImageName, image),
			steps.CreateContainer(p.nodeClients, createReq, &contId),
			steps.AssembleConfigStep(p.nodeClients, p.services, &contId, createReq, image, res),
			steps.DropContainerStep(p.nodeClients, &contId),
		},

		getResult: func() (*matreshka.AppConfig, error) {
			return res, nil
		},
	}
}
