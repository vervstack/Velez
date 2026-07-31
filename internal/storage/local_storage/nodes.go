package local_storage

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/hardware"
	"go.vervstack.ru/Velez/internal/domain"
)

type nodes struct {
	containerAPI node_clients.Docker
	region       string
}

func newNodesStorage(containerAPI node_clients.Docker, region string) *nodes {
	return &nodes{containerAPI: containerAPI, region: region}
}

func (n *nodes) InitNode(_ context.Context, _ string) error {
	return nil
}

func (n *nodes) UpdateOnline(_ context.Context, _, _ float64) error {
	return nil
}

func (n *nodes) List(ctx context.Context, _ domain.ListNodesReq) (domain.NodesList, error) {
	cpuPercent, memPercent, err := hardware.New("").GetUsage(ctx)
	if err != nil {
		return domain.NodesList{}, rerrors.Wrap(err, "error getting node usage")
	}

	servicesCount, err := countRunningServices(ctx, n.containerAPI)
	if err != nil {
		return domain.NodesList{}, rerrors.Wrap(err, "error counting running services")
	}

	return domain.NodesList{
		Nodes: []domain.NodeBaseInfo{
			{
				Id:            -1,
				Name:          "VelezSingleNode",
				LastOnline:    time.Now(),
				Addr:          "",
				IsEnabled:     true,
				CpuPercent:    cpuPercent,
				MemPercent:    memPercent,
				ServicesCount: servicesCount,
				Region:        n.region,
			},
		},
		Total: 1,
	}, nil
}
