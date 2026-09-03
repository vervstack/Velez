package verv_services

import (
	"context"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/utils/configutils"
)

// reconcileResources runs docs/features/vervonomicon.md's "Resource
// reconciliation" over a descriptor's resources[]: existing
// velez.service_resources bindings come from storage, the service's live
// connections come from its matreshka config (the same access path
// Configurator.GetVervFromApi already provides for other config reads) - the
// actual host-vs-placeholder decision is pure logic in
// vervonomicon.ReconcileResources.
func (v *VervService) reconcileResources(
	ctx context.Context, serviceName string, resources []verv.Resource,
) ([]domain.ResourceReconciliation, error) {
	if len(resources) == 0 {
		return nil, nil
	}

	bindings, err := v.dataStorage.ServiceResources().GetResources(ctx, serviceName)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting existing service resource bindings")
	}

	meta := domain.ConfigMeta{
		Name:     configutils.AppendPrefix(matreshka_api.ConfigTypePrefix_verv, serviceName),
		ConfType: matreshka_api.ConfigTypePrefix_verv,
		Format:   velez_api.ConfigFormat_env,
	}

	// A matreshka config that cannot be read is not a reconciliation failure.
	// matreshka is one of two independent signals here (the other being the
	// service_resources bindings above), and its absence carries the same
	// meaning as an entry with a placeholder host: no live connection is
	// known. Hard-failing here would take GetVervonomicon down with matreshka
	// - and a service that never had a matreshka config at all could never be
	// reconciled.
	var hosts map[string]string

	appConfig, err := v.configService.GetVervFromApi(ctx, meta)
	if err != nil {
		log.Ctx(ctx).Warn().
			Str("service_name", serviceName).
			Err(err).
			Msg("could not read matreshka config for resource reconciliation, continuing on bindings alone")
	} else {
		hosts = vervonomicon.DataSourceHosts(appConfig)
	}

	return vervonomicon.ReconcileResources(resources, bindings, hosts), nil
}
