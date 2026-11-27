package jobs

import (
	"context"

	"go.redsock.ru/rerrors"
)

const (
	DropVcnNetworkAction = "drop_vcn_network"
)

type DropVcnNetworkPayload struct {
	NetworkId string `json:"network_id"`
	NodeName  string `json:"node_name"`
}

type VcnNetworkDropper interface {
	DropLocal(ctx context.Context, id string) error
}

type dropVcnNetworkHandler struct {
	networks VcnNetworkDropper
}

func NewDropVcnNetworkHandler(networks VcnNetworkDropper) TaskHandler {
	return &dropVcnNetworkHandler{networks: networks}
}

func (h *dropVcnNetworkHandler) Action() string {
	return DropVcnNetworkAction
}

func (h *dropVcnNetworkHandler) NewContext() TaskContext {
	return &DropVcnNetworkPayload{}
}

func (h *dropVcnNetworkHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*DropVcnNetworkPayload)
	if !ok {
		panic("drop_vcn_network: BuildJobs called with mismatched TaskContext type")
	}

	return []NamedJob{
		{Name: "remove_network", Job: &dropVcnNetworkJob{networks: h.networks, id: payload.NetworkId}},
	}
}

type dropVcnNetworkJob struct {
	networks VcnNetworkDropper
	id       string
}

func (j *dropVcnNetworkJob) Do(ctx context.Context) error {
	err := j.networks.DropLocal(ctx, j.id)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}
