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
	Id        int64
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
	Id        int64
	Name      *string
	Type      *RegistryType
	Url       *string
	Username  *string
	Secret    *string
	IsDefault *bool
}

type DeleteRegistryReq struct {
	Id   *int64
	Name *string
}
