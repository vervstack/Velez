package deploy_steps

import (
	"context"

	"github.com/godverv/Velez/pkg/velez_api"
)

type prepareRequestStep struct {
	req *velez_api.CreateSmerd_Request
}

func PrepareRequest(req *velez_api.CreateSmerd_Request) *prepareRequestStep {
	return &prepareRequestStep{
		req: req,
	}
}

func (p *prepareRequestStep) Do(_ context.Context) error {
	if p.req.Settings == nil {
		p.req.Settings = &velez_api.Container_Settings{}
	}

	if p.req.Hardware == nil {
		p.req.Hardware = &velez_api.Container_Hardware{}
	}

	if p.req.Env == nil {
		p.req.Env = make(map[string]string)
	}

	if p.req.Labels == nil {
		p.req.Labels = make(map[string]string)
	}

	return nil
}
