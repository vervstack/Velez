package network_steps

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/makosh/pkg/makosh_be"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

type addMakoshRecord struct {
	sd          cluster_clients.ServiceDiscovery
	serviceName string
	vcnAddr     []string // Verv Closed Network address (both ip and hostname
}

func AddMakoshRecord(
	sd cluster_clients.ServiceDiscovery,
	serviceName string,
	vcnAddrs ...string,
) steps.Step {
	return &addMakoshRecord{
		sd:          sd,
		serviceName: serviceName,
		vcnAddr:     vcnAddrs,
	}
}

func (a *addMakoshRecord) Do(ctx context.Context) error {

	upsertReq := &makosh_be.UpsertEndpoints_Request{
		Endpoints: []*makosh_be.Endpoint{
			{
				ServiceName: a.serviceName,
				Addrs:       a.vcnAddr,
			},
		},
	}

	_, err := a.sd.UpsertEndpoints(ctx, upsertReq)
	if err != nil {
		return rerrors.Wrap(err, "error during upsertion of makosh endpoints")
	}

	return nil
}
