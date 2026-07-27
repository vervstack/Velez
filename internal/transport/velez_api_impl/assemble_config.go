package velez_api_impl

import (
	"context"
	"encoding/json"
	"time"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

// assembleConfigWatchTimeout bounds how long the synchronous AssembleConfig
// RPC blocks waiting for the assemble_config task to reach a terminal
// status. Same safety-net rationale as createSmerdWatchTimeout/
// createServiceWatchTimeout - assemble_config additionally pulls an image
// and spins up/tears down a scratch container, so it's given the same
// budget as create_smerd rather than create_service's lighter DB-only one.
const (
	assembleConfigWatchTimeout = 60 * time.Second
)

func (impl *Impl) AssembleConfig(ctx context.Context, req *velez_api.AssembleConfig_Request) (
	*velez_api.AssembleConfig_Response, error) {
	initialContext := &velez_api.AssembleConfigTaskPayload{
		ServiceName: req.GetServiceName(),
		ImageName:   req.GetImageName(),
	}

	_, err := impl.jobsEngine.Enqueue(ctx, req.GetServiceName(), jobs.AssembleConfigAction, initialContext)
	if err != nil {
		return nil, rerrors.Wrap(err, "error enqueuing assemble_config task")
	}

	watchCtx, cancel := context.WithTimeout(ctx, assembleConfigWatchTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	for task := range impl.jobsEngine.Watch(watchCtx, req.GetServiceName(), jobs.AssembleConfigAction) {
		finalTask = task
	}

	if finalTask.Status != tasks_queries.VelezTaskStatusDONE &&
		finalTask.Status != tasks_queries.VelezTaskStatusFAILED &&
		watchCtx.Err() != nil {
		return nil, rerrors.Wrapf(
			watchCtx.Err(), "timed out waiting for assemble_config task, last status: %q", finalTask.Status,
		)
	}

	if finalTask.Status == tasks_queries.VelezTaskStatusFAILED {
		return nil, rerrors.New(finalTask.Error.String)
	}

	payload := &velez_api.AssembleConfigTaskPayload{}

	err = json.Unmarshal(finalTask.Context.RawMessage, payload)
	if err != nil {
		return nil, rerrors.Wrap(err, "error unmarshaling task context")
	}

	confType := matreshka_api.ConfigTypePrefix(matreshka_api.ConfigTypePrefix_value[payload.GetConfType()])

	appConfig := domain.AppConfig{
		Meta: domain.ConfigMeta{
			Name:     payload.GetConfigName(),
			Version:  payload.ConfigVersion,
			ConfType: confType,
			Format:   payload.GetConfigFormat(),
		},
		ContentRaw: payload.GetContentRaw(),
	}

	err = impl.cfgService.UpdateConfig(ctx, appConfig)
	if err != nil {
		if !rerrors.Is(err, cluster_clients.ErrServiceIsDisabled) {
			return nil, rerrors.Wrap(err, "error updating config")
		}
	}

	return &velez_api.AssembleConfig_Response{
		Config: appConfig.ContentRaw,
	}, nil
}
