package domain

import (
	"time"
)

type NodeType int

const (
	NodeTypeService  NodeType = 0
	NodeTypeResource NodeType = 1
)

type ServiceDependency struct {
	SourceService string
	TargetService string
	Proto         string
	RequestRate   float64
	NodeType      NodeType
}

type ServiceGraph struct {
	Callers      []ServiceDependency
	Dependencies []ServiceDependency
}

type ServiceEnvironment struct {
	Env             string
	Status          string
	DeployedVersion string
	DeployedAt      *time.Time
	Health          string
}
