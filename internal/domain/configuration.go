package domain

import (
	"go.redsock.ru/evon"
	"go.vervstack.ru/matreshka/pkg/matreshka_be_api"
)

type ConfigurationPatch struct {
	ConfigName string
	EnvVarsMap map[string]*string
}

type ConfigMeta struct {
	Name     string
	Version  *string
	ConfType matreshka_be_api.ConfigTypePrefix
}

type AppConfig struct {
	Meta    ConfigMeta
	Content *evon.Node
}
