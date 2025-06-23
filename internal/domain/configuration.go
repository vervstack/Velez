package domain

import (
	"go.redsock.ru/evon"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"
)

type ConfigurationPatch struct {
	ConfigName string
	EnvVarsMap map[string]*string
}

type ConfigMeta struct {
	Name     string
	Version  *string
	ConfType matreshka_api.ConfigTypePrefix
	Format   matreshka_api.Format
}

type AppConfig struct {
	Meta       ConfigMeta
	Content    *evon.Node
	ContentRaw []byte
}

type ConfigMount struct {
	Meta ConfigMeta

	FilePath *string
	Content  []byte
}
