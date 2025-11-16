package pipelines

import (
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (p *pipeliner) EnableVervService(req velez_api.VervServiceType) Runner[any] {

	return &runner[any]{
		Steps: []steps.Step{
			// Enable service in Velez's config
			// Start container
		},
	}
}
