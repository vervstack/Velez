package domain

type SysboxStatus struct {
	OsType              string
	KernelVersion       string
	DockerVersion       string
	IsRootless          bool
	IsSnap              bool
	IsRuntimeRegistered bool
	ContainersTotal     int32
	ContainersOnSysbox  int32
}

type SysboxSmokeTestResult struct {
	IsPassed bool
	Failure  string
}
