package steps

import (
	"context"

	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type enableVervServiceInMatreshkaStep struct {
	cfg service.ConfigurationService

	vervServiceType velez_api.VervServiceType
}

func EnableVervServiceInMatreshka(cfg service.Services, vervServiceType velez_api.VervServiceType) Step {
	return &enableVervServiceInMatreshkaStep{
		cfg:             cfg.ConfigurationService(),
		vervServiceType: vervServiceType,
	}
}

func (p *enableVervServiceInMatreshkaStep) Do(ctx context.Context) (err error) {

	return nil
}

func (p *enableVervServiceInMatreshkaStep) Rollback(_ context.Context) error {

	return nil
}
