package domain

import (
	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

type PluginBaseInfo struct {
	Name           string
	State          pb.VervPlugin_State
	ServiceId      *int64
	DeployStatuses []pb.DeploymentStatus
}
