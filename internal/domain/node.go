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

	CpuPercent    float64
	MemPercent    float64
	ServicesCount uint64

	Region string
}
