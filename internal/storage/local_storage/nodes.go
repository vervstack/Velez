package local_storage

import (
	"context"
	"time"

	"go.vervstack.ru/Velez/internal/domain"
)

type nodes struct{}

func newNodesStorage() *nodes {
	return &nodes{}
}

func (n *nodes) InitNode(_ context.Context) error {
	return nil
}

func (n *nodes) UpdateOnline(_ context.Context) error {
	return nil
}

func (n *nodes) List(_ context.Context, _ domain.ListNodesReq) (domain.NodesList, error) {
	return domain.NodesList{
		Nodes: []domain.NodeBaseInfo{
			{
				Id:         -1,
				Name:       "VelezSingleNode",
				LastOnline: time.Now(),
				Addr:       "",
				IsEnabled:  true,
			},
		},
		Total: 1,
	}, nil
}
