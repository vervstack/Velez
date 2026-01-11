package config_mocks

import (
	_ "embed"
)

var (
	//go:embed loki.yaml
	Loki []byte

	//go:embed test_config.yaml
	DefaultTestConfig []byte

	//go:embed hello_world.yaml
	HelloWorld []byte
)
