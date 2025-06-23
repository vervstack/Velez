package configurator

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/utils/configutils"
)

func (c *Configurator) GetPlainFromApi(ctx context.Context, meta domain.ConfigMeta) ([]byte, error) {
	getReq := &matreshka_api.GetConfig_Request{
		ConfigName: configutils.AppendPrefix(meta.ConfType, meta.Name),
		Version:    meta.Version,
		Format:     meta.Format,
	}

	cfg, err := c.MatreshkaBeAPIClient.GetConfig(ctx, getReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting config")
	}

	return cfg.Config, nil
}
