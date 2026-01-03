package domain

type CreateServiceReq struct {
	Name string
}

type Service struct {
	Id uint64
	ServiceBasicInfo
}

type ServiceBasicInfo struct {
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
