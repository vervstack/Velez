package config_mocks

import (
	_ "embed"
)

var (
	//go:embed loki.yaml
	Loki []byte

	//go:embed hello_world.yaml
	HelloWorld []byte
)
