package pipelines

import (
	"github.com/docker/docker/api/types/image"
	"go.redsock.ru/evon"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

const configFetchingPostfix = "_config_scanning"

func (p *pipeliner) AssembleConfig(req domain.AssembleConfig) Runner[domain.AppConfig] {
	imageResp := &image.InspectResponse{}
	contId := ""

	createReq := domain.LaunchSmerd{
		CreateSmerd_Request: &velez_api.CreateSmerd_Request{
			Name:      req.ServiceName + configFetchingPostfix,
			ImageName: req.ImageName,
			Settings:  &velez_api.Container_Settings{},
		}}

	res := &domain.AppConfig{
		Meta: domain.ConfigMeta{
			Name: req.ServiceName,
		},
		Content: &evon.Node{},
	}

	return &runner[domain.AppConfig]{
		Steps: []steps.Step{
			steps.PrepareImageStep(p.nodeClients, req.ImageName, imageResp),
			steps.CreateContainer(p.nodeClients, createReq, &contId),
			steps.AssembleConfigStep(p.nodeClients, p.services, &contId, createReq, imageResp, res),
			steps.DropContainerStep(p.nodeClients, &contId),
		},
		getResult: func() (*domain.AppConfig, error) {
			return res, nil
		},
	}
}
