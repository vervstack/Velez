package pipelines

import (
	"github.com/docker/docker/api/types"
	"go.vervstack.ru/matreshka"

	"github.com/godverv/Velez/internal/pipelines/steps"
	"github.com/godverv/Velez/pkg/velez_api"
)

const configFetchingPostfix = "_config_scanning"

func (p *pipeliner) GetConfig(req *velez_api.FetchConfig_Request) Runner[matreshka.AppConfig] {
	image := &types.ImageInspect{}
	contId := ""
	createReq := &velez_api.CreateSmerd_Request{
		Name:      req.ServiceName + configFetchingPostfix,
		ImageName: req.ImageName,
		Settings:  &velez_api.Container_Settings{},
	}

	res := &matreshka.AppConfig{}

	return &runner[matreshka.AppConfig]{
		Steps: []PipelineStep{
			steps.PrepareImageStep(p.dockerAPI, req.ImageName, image),
			steps.LaunchContainer(p.dockerAPI, createReq, &contId),
			steps.AssembleConfigStep(p.dockerAPI, p.configService, &contId, req.ServiceName, res),
		},

		getResult: func() (*matreshka.AppConfig, error) {
			return res, nil
		},
	}
}
