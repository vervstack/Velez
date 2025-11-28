package domain

import (
	"go.redsock.ru/evon"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

type ConfigurationPatch struct {
	ConfigName string
	EnvVarsMap map[string]*string
}

type ConfigMeta struct {
	Name     string
	Version  *string
	ConfType matreshka_api.ConfigTypePrefix
	Format   velez_api.ConfigFormat
}

type AppConfig struct {
	Meta       ConfigMeta
	Content    *evon.Node
	ContentRaw []byte
}

type ConfigMount struct {
	Meta ConfigMeta
	FileMountPoint
}

type FileMountPoint struct {
	FilePath *string
	Content  []byte
}
