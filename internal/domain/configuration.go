package domain

type ConfigurationPatch struct {
	ServiceName string
	EnvVarsMap  map[string]*string
}
