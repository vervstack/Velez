package pipelines

import (
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/assert/yaml"
	"go.redsock.ru/evon"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"

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

	configMount := &domain.ConfigMount{}

	return &runner[domain.AppConfig]{
		Steps: []steps.Step{
			steps.PrepareImage(p.nodeClients, req.ImageName, imageResp),
			steps.CreateContainer(p.nodeClients, &createReq, &contId),
			steps.GetConfigFromContainerStep(p.nodeClients, p.services, &createReq, &contId, imageResp, configMount),
			steps.DropContainerStep(p.nodeClients, &contId),
		},
		getResult: func() (*domain.AppConfig, error) {
			configMount.Meta.Name = strings.ReplaceAll(configMount.Meta.Name, configFetchingPostfix, "")
			appConfig := domain.AppConfig{
				Meta:       configMount.Meta,
				ContentRaw: configMount.Content,
			}
			switch appConfig.Meta.Format {
			case matreshka_api.Format_yaml:
				m := map[string]any{}
				err := yaml.Unmarshal(configMount.Content, &m)
				if err != nil {
					return nil, rerrors.Wrap(err, "error unmarshalling from yaml to map")
				}

				appConfig.Content, err = evon.MarshalEnv(m)
				if err != nil {
					return nil, rerrors.Wrap(err, "error marshalling yaml to env")
				}

			case matreshka_api.Format_env:
				err := evon.Unmarshal(configMount.Content, &appConfig.Content)
				if err != nil {
					return nil, rerrors.Wrap(err, "error unmarshalling to evon ")
				}
			}

			return &appConfig, nil
		},
	}
}
