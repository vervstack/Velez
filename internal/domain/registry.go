package domain

import (
	"time"
)

type RegistryType string

const (
	RegistryTypeDockerHub RegistryType = "dockerhub"
	RegistryTypeGenericV2 RegistryType = "generic_v2"
)

type Registry struct {
	ID        int64
	Name      string
	Type      RegistryType
	Url       string
	Username  string
	Secret    string
	IsDefault bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateRegistryReq struct {
	Name      string
	Type      RegistryType
	Url       string
	Username  string
	Secret    string
	IsDefault bool
}

type UpdateRegistryReq struct {
	ID        int64
	Name      *string
	Type      *RegistryType
	Url       *string
	Username  *string
	Secret    *string
	IsDefault *bool
}

type DeleteRegistryReq struct {
	ID   *int64
	Name *string
}
