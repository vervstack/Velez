package pipelines

import (
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/assert/yaml"
	"go.redsock.ru/evon"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
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
		getResult: func() (appConfig *domain.AppConfig, err error) {
			configMount.Meta.Name = strings.ReplaceAll(configMount.Meta.Name, configFetchingPostfix, "")
			appConfig = &domain.AppConfig{
				Meta:       configMount.Meta,
				ContentRaw: configMount.Content,
			}
			switch appConfig.Meta.Format {
			case matreshka_api.Format_yaml:
				if labels.IsMatreshkaImage(imageResp) {
					appConfig.Content, err = fromMatreshkaYamlToEvon(configMount.Content)
				} else {
					appConfig.Content, err = fromYamlToEvon(configMount.Content)
				}
				if err != nil {
					return nil, rerrors.Wrap(err, "error parsing from yaml to evon")
				}
			case matreshka_api.Format_env:
				err := evon.Unmarshal(configMount.Content, &appConfig.Content)
				if err != nil {
					return nil, rerrors.Wrap(err, "error unmarshalling to evon ")
				}
			}

			return appConfig, nil
		},
	}
}

func fromMatreshkaYamlToEvon(content []byte) (*evon.Node, error) {
	c := matreshka.NewEmptyConfig()
	err := c.Unmarshal(content)
	if err != nil {
		return nil, rerrors.Wrap(err, "error unmarshalling to matreshka config")
	}

	env, err := evon.MarshalEnv(c)
	if err != nil {
		return nil, rerrors.Wrap(err, "error marshalling to env")
	}

	return env, nil
}

func fromYamlToEvon(content []byte) (*evon.Node, error) {
	m := map[string]any{}
	err := yaml.Unmarshal(content, &m)
	if err != nil {
		return nil, rerrors.Wrap(err, "error unmarshalling from yaml to map")
	}

	env, err := evon.MarshalEnv(m)
	if err != nil {
		return nil, rerrors.Wrap(err, "error marshalling yaml to env")
	}

	return env, nil
}
