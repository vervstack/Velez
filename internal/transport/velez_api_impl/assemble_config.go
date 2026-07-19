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
const assembleConfigWatchTimeout = 60 * time.Second

func (impl *Impl) AssembleConfig(ctx context.Context, req *velez_api.AssembleConfig_Request) (*velez_api.AssembleConfig_Response, error) {
	initialContext := &velez_api.AssembleConfigTaskPayload{
		ServiceName: req.ServiceName,
		ImageName:   req.ImageName,
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

	isDone := finalTask.Status == tasks_queries.VelezTaskStatusDONE
	isFailed := finalTask.Status == tasks_queries.VelezTaskStatusFAILED

	if !isDone && !isFailed && watchCtx.Err() != nil {
		return nil, rerrors.Wrapf(
			watchCtx.Err(),
			"timed out waiting for assemble_config task, last status: %q",
			finalTask.Status,
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

	// Reconstruct the domain.AppConfig UpdateConfig needs. Meta.Name/
	// Version/ConfType/Format come straight from the persisted payload,
	// converting the payload's plain-string conf_type (kept as a string in
	// the proto to avoid a matreshka_api proto dependency - see
	// AssembleConfigTaskPayload.conf_type's comment) into the typed enum via
	// matreshka_api.ConfigTypePrefix_value.
	//
	// Meta.Content (*evon.Node) is deliberately left nil rather than
	// reconstructed from the payload's content field. The old pipeline's own
	// getResult, for the ConfigFormat_env case (which is the only format
	// AssembleConfig ever produces - see classifyImage/fillMeta, neither
	// sets a Verv/Plain override with a .yaml FilePath), did
	// `evon.Unmarshal(bytes, &appConfig.Content)` where Content is a bare
	// *evon.Node. That call is a silent no-op: evon.Unmarshal only knows how
	// to populate struct/map destinations, and reflect.Value.Elem() on a nil
	// *evon.Node target yields an invalid Value, so no field mapping is ever
	// built and Content is left nil. Verified empirically (throwaway
	// evon.Marshal/Unmarshal round-trip) rather than just read off the
	// reflect code. So today, in production, AppConfig.Content is already
	// always nil by the time UpdateConfig runs - this cutover reproduces
	// that exactly rather than "fixing" a pre-existing bug as a side effect.
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

	resp := velez_api.AssembleConfig_Response{
		Config: appConfig.ContentRaw,
	}

	return &resp, nil
}
