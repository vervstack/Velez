package domain

type ConfigurationPatch struct {
	ServiceName string
	EnvVarsMap  map[string]*string
}

type ConfigMeta struct {
	ServiceName string
	CfgVersion  *string
}
