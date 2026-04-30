package steps

import (
	"context"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

type prepareRequestStep struct {
	req *domain.LaunchSmerd
}

func PrepareCreateRequest(req *domain.LaunchSmerd) *prepareRequestStep {
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

	if p.req.Config == nil {
		p.req.Config = &velez_api.CreateSmerd_Request_Verv{
			Verv: &velez_api.MatreshkaConfigSpec{
				ConfigName: &p.req.Name,
			},
		}
	}

	return nil
}
