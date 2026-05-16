package postgres

import (
	"context"
	"database/sql"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/plugins_queries"
)

type pluginsStorage struct {
	querier plugins_queries.Querier
}

func newPluginsStorage(db *sql.DB) *pluginsStorage {
	return &pluginsStorage{
		querier: plugins_queries.New(db),
	}
}

func (p *pluginsStorage) ListPlugins(ctx context.Context) ([]domain.PluginBaseInfo, error) {
	plugins, err := p.querier.ListPlugins(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing plugins")
	}

	result := make([]domain.PluginBaseInfo, 0, len(plugins))
	for _, pluginRow := range plugins {

		plugin := domain.PluginBaseInfo{
			Name:  pluginRow.PluginType,
			State: 0,
		}

		if pluginRow.ServiceID.Valid {
			plugin.ServiceId = &pluginRow.ServiceID.Int64
		}

		plugin.State = calculatePluginState(pluginRow.Statuses)

		result = append(result, plugin)
	}

	return result, nil
}

func (p *pluginsStorage) UpsertPlugin(ctx context.Context, arg plugins_queries.UpsertPluginParams) error {
	return p.querier.UpsertPlugin(ctx, arg)
}

func calculatePluginState(statusesArr []string) pb.VervPlugin_State {
	isRunning := false
	isDeleted := false
	for _, status := range statusesArr {
		switch plugins_queries.VelezDeploymentStatus(status) {
		case plugins_queries.VelezDeploymentStatusFAILED:
			return pb.VervPlugin_dead

		case plugins_queries.VelezDeploymentStatusRUNNING:
			isRunning = true

		case plugins_queries.VelezDeploymentStatusDELETED:
			isDeleted = true

		case plugins_queries.VelezDeploymentStatusSCHEDULEDUPGRADE,
			plugins_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT:
			return pb.VervPlugin_warning
		}
	}

	if isRunning {
		return pb.VervPlugin_running
	}

	if isDeleted {
		return pb.VervPlugin_disabled
	}

	return pb.VervPlugin_unknown
}
