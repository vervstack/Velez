package domain

import (
	"time"
)

// DindInstance is the dind-specific satellite row (velez.dind_instances) for a
// Docker-in-Docker service.
type DindInstance struct {
	ServiceId       int64
	IsSysboxEnabled bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type UpsertDindInstanceReq struct {
	ServiceId       int64
	IsSysboxEnabled bool
}

type CreateDindReq struct {
	Name            string
	Environment     string
	IsSysboxEnabled bool
}

type DindView struct {
	DindInstance

	Name    string
	Address string
}
