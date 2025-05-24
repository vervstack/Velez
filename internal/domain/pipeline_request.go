package domain

import (
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type LaunchSmerd struct {
	*velez_api.CreateSmerd_Request
}

type LaunchSmerdResult struct {
	ContainerId string
}
