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
}

type GetServiceReq struct {
	Name string
}

type RemoveServiceReq struct {
	Name                 string
	DropRunningInstances bool
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
}

type ServiceList struct {
	Total    uint64
	Services []ServiceBaseInfo
}
