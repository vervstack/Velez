package configurator

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/utils/configutils"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"
)

func (c *Configurator) GetPlainFromApi(ctx context.Context, meta domain.ConfigMeta) ([]byte, error) {
	getReq := &matreshka_api.GetConfig_Request{
		ConfigName: configutils.AppendPrefix(meta.ConfType, meta.Name),
		Version:    meta.Version,
		Format:     matreshka_api.Format(meta.Format),
	}

	cfg, err := c.GetConfig(ctx, getReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting config")
	}

	return cfg.GetConfig(), nil
}
