package domain

import (
	"time"

	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
)

type ListDeploymentsReq struct {
	ServiceIds []int64
	NodeIds    []int64
	NotStatus  []deployments_queries.VelezDeploymentStatus

	Paging Paging
}
type Deployment struct {
	Id        int64
	ServiceId int64
	SpecId    int64
	NodeId    int64

	CreatedAt time.Time
	UpdatedAt time.Time
	Status    deployments_queries.VelezDeploymentStatus
}

type DeploymentSpecification struct {
	Id      int64
	Payload []byte
}
