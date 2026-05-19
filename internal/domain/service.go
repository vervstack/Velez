package domain

import (
	"time"

	rtb "go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

type CreateServiceReq struct {
	Name string
}

type Service struct {
	ServiceBaseInfo
	CurrentDeploymentId *uint64
	Status              velez_api.DeploymentStatus
}

type ServiceBaseInfo struct {
	Name           string
	LastDeployedAt *time.Time
	ImageName      string
	Status         string
	Env            string
}

type GetServiceReq struct {
	Name string
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
