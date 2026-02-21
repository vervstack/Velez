package domain

import (
	"github.com/docker/docker/api/types/container"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"
)

type EnableStatefullClusterRequest struct {
	ExposePort   bool
	ExposeToPort uint64
}

type StateClusterDefinition struct {
	CreateReq    container.CreateRequest
	RootPostgres resources.Postgres
}
