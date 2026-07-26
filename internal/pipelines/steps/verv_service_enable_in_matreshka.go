package steps

import (
	"context"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/service"
)

type enableVervServiceInMatreshkaStep struct {
	cfg service.ConfigurationService

	vervServiceType velez_api.VervPluginType
}

func EnableVervServiceInMatreshka(cfg service.Services, vervServiceType velez_api.VervPluginType) Step {
	return &enableVervServiceInMatreshkaStep{
		cfg:             cfg.ConfigurationService(),
		vervServiceType: vervServiceType,
	}
}

func (p *enableVervServiceInMatreshkaStep) Do(_ context.Context) (err error) {
	return nil
}

func (p *enableVervServiceInMatreshkaStep) Rollback(_ context.Context) error {
	return nil
}
