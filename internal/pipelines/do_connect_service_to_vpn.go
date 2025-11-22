package pipelines

import (
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

func (p *pipeliner) ConnectServiceToVpn(req domain.ConnectServiceToVpn) Runner[any] {

	return &runner[any]{
		Steps: []steps.Step{},
	}
}
