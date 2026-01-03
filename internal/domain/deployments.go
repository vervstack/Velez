package domain

import (
	"time"
)

type ListDeploymentsReq struct {
	ServiceId int64
}
type Deployment struct {
	Id        int64
	ServiceId int64
	CreatedAt time.Time
}
