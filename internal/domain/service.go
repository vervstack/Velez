package domain

import (
	"time"

	rtb "go.redsock.ru/toolbox"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

type CreateServiceReq struct {
	Name string
}

type AboutService struct {
	Description  string
	OriginalName string
	Env          string
	ServiceType  string
	Team         string
	Repo         string
	Port         string
}

type Service struct {
	ServiceBaseInfo

	ID int64

	CurrentDeploymentId *uint64
	Status              velez_api.DeploymentStatus
	About               AboutService
}

type ServiceBaseInfo struct {
	Name           string
	LastDeployedAt *time.Time
	ImageName      string
	Status         string
	Env            string
	Repo           string
	// Labels - derived, never stored. One of "service-core", "service-app",
	// "resource-<type>". Populated by the storage layer via ClassifyService.
	Labels []string
}

type GetServiceReq struct {
	Name string
}

type RemoveServiceReq struct {
	Name                 string
	DropRunningInstances bool

	// Environment - the isolated namespace this operation targets. Empty
	// means the default/PROD environment (see storage/environments.Resolve).
	Environment string
}

type CreateDeployReq struct {
	LaunchSmerd

	ServiceName string
}

type UpgradeDeployReq struct {
	ServiceName  string
	DeploymentId uint64

	NewImage *string
}

type ListServicesReq struct {
	Paging      Paging
	NamePattern rtb.Optional[string]

	// IncludeInternal - when false, Verv-internal entries (core services and
	// anything bound as a resource) are excluded from the result.
	IncludeInternal bool
}

type ServiceList struct {
	Total    uint64
	Services []ServiceBaseInfo
}
