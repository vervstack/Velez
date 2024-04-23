package parser

import (
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"

	"github.com/godverv/Velez/pkg/velez_api"
)

func FromCommand(command *string) strslice.StrSlice {
	if command == nil {
		return nil
	}

	return strings.Split(*command, " ")
}

func FromHealthcheck(healthcheck *velez_api.Container_Healthcheck) *container.HealthConfig {
	if healthcheck == nil {
		return nil
	}

	return &container.HealthConfig{
		Test: []string{
			"CMD-SHELL",
			healthcheck.GetCommand(),
		},
		Interval: time.Second * time.Duration(healthcheck.GetIntervalSecond()),
		Timeout:  time.Second * time.Duration(healthcheck.GetTimeoutSecond()),
		Retries:  int(healthcheck.GetRetries()),
	}
}
