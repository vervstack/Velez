package service

import (
	"context"

	"go.vervstack.ru/Velez/internal/domain"
)

type VcnService interface {
	Create(ctx context.Context, req CreateVcnNetworkRequest) (domain.Network, error)
	List(ctx context.Context) ([]domain.Network, error)
	Get(ctx context.Context, id string) (VcnNetworkDetails, error)
	Drop(ctx context.Context, id string) error
	Peers(ctx context.Context, id string) ([]domain.VcnNode, error)
	IssueApiKey(ctx context.Context, id string) (string, error)
	ExpireApiKey(ctx context.Context, id, prefix string) error
	CurrentNode(ctx context.Context) (string, error)
}

type CreateVcnNetworkRequest struct {
	Name string
	Host string

	Deploy   *domain.SetupHeadscaleRequest
	External *VcnExternalSource
}

type VcnExternalSource struct {
	Url   string
	Token string
}

type VcnNetworkDetails struct {
	Network domain.Network
	Peers   []domain.VcnNode
	ApiKeys []domain.VcnApiKey
}
