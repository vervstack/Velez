package domain

import "time"

type ServiceDependency struct {
	SourceService string
	TargetService string
	Proto         string
	RequestRate   float64
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
