package domain

import (
	"time"

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

type ListNodesReq struct {
	Paging Paging
}
type NodesList struct {
	Nodes []NodeBaseInfo
	Total uint64
}

type NodeBaseInfo struct {
	Id         int64
	Name       string
	LastOnline time.Time
	Addr       string
	IsEnabled  bool
}
