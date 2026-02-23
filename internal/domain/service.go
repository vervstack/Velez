package domain

import (
	rtb "go.redsock.ru/toolbox"
)

type CreateServiceReq struct {
	Name string
}

type Service struct {
	ServiceBaseInfo
}

type ServiceBaseInfo struct {
	Id   uint64
	Name string
}

type GetServiceReq struct {
	Id   *uint64
	Name *string
}

type CreateDeployReq struct {
	LaunchSmerd
	ServiceId uint64
}

type UpgradeDeployReq struct {
	ServiceId    uint64
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
