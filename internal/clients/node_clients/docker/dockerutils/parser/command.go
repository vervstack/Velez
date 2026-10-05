package parser

import (
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

const (
	healthcheckTestExec  = "CMD"
	healthcheckTestShell = "CMD-SHELL"
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

	var test []string

	switch {
	case len(healthcheck.GetExec()) > 0:
		test = append([]string{healthcheckTestExec}, healthcheck.GetExec()...)
	case healthcheck.GetCommand() != "":
		test = []string{healthcheckTestShell, healthcheck.GetCommand()}
	default:
		return nil
	}

	return &container.HealthConfig{
		Test:     test,
		Interval: time.Second * time.Duration(healthcheck.GetIntervalSecond()),
		Timeout:  time.Second * time.Duration(healthcheck.GetTimeoutSecond()),
		Retries:  int(healthcheck.GetRetries()),
	}
}
