package domain

import (
	"time"

	"github.com/google/uuid"

	"go.vervstack.ru/Velez/internal/storage/postgres/generated/transfers_queries"
)

type Transfer struct {
	Uuid          uuid.UUID
	IsSource      bool
	ContainerName string
	PeerAddress   string
	Status        transfers_queries.VelezTransferStatus
	Error         string
	TotalBytes    int64
	DoneBytes     int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TransferPlan struct {
	Sidecars          []TransferSidecar
	Consumers         []TransferConsumer
	UnsupportedMounts []string
}

type TransferSidecar struct {
	ContainerName string
	ImageName     string
}

type TransferConsumer struct {
	ContainerName string
	NetworkName   string
}

type StartTransferReq struct {
	ContainerName           string
	SidecarNames            []string
	Address                 string
	Token                   string
	IsConsumersAcknowledged bool
}
