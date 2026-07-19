package domain

import (
	"time"
)

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
