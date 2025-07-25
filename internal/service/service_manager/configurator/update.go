package configurator

import (
	"context"
	"fmt"

	"go.redsock.ru/evon"
	errors "go.redsock.ru/rerrors"
	matrapi "go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/utils/configutils"
)

func (c *Configurator) UpdateConfig(ctx context.Context, cfg domain.AppConfig) (err error) {
	patchRequest := &matrapi.PatchConfig_Request{
		Version: cfg.Meta.Version,
	}

	patchRequest.ConfigName = configutils.AppendPrefix(cfg.Meta.ConfType, cfg.Meta.Name)

	oldCfg, err := c.getEnvFromApi(ctx, domain.ConfigMeta{Name: patchRequest.ConfigName})
	if err != nil {
		return errors.Wrap(err, "error getting actual config")
	}

	diff := evon.Diff(oldCfg, cfg.Content)

	patchRequest.Patches = make([]*matrapi.PatchConfig_Patch, 0, len(diff.NewNodes)+len(diff.RemovedNodes))

	for _, d := range diff.NewNodes {
		if d.Value == nil {
			continue
		}

		patchRequest.Patches = append(patchRequest.Patches,
			&matrapi.PatchConfig_Patch{
				FieldName: d.Name,
				Patch: &matrapi.PatchConfig_Patch_UpdateValue{
					UpdateValue: fmt.Sprint(d.Value),
				},
			},
		)
	}

	for _, d := range diff.RemovedNodes {
		patchRequest.Patches = append(patchRequest.Patches,
			&matrapi.PatchConfig_Patch{
				FieldName: d.Name,
				Patch: &matrapi.PatchConfig_Patch_Delete{
					Delete: true,
				},
			},
		)
	}

	_, err = c.MatreshkaBeAPIClient.PatchConfig(ctx, patchRequest)
	if err != nil {
		return errors.Wrap(err, "error patching config")
	}

	return nil
}
