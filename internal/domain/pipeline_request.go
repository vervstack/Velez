package domain

import (
	velez_api "go.vervstack.ru/Velez/internal/api/server/api/grpc"
)

type LaunchSmerd struct {
	*velez_api.CreateSmerd_Request
}

type LaunchSmerdResult struct {
	ContainerId string
}
