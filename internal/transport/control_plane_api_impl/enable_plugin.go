package control_plane_api_impl

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/state"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

var (
	errUnsupportedService = rerrors.New("unsupported service", codes.InvalidArgument)
)

// enableStatefullWatchTimeout bounds how long the synchronous EnablePlugin
// RPC blocks waiting for the enable_statefull_mode task to reach a terminal
// status. Same safety-net rationale as createSmerdWatchTimeout/
// assembleConfigWatchTimeout/connectServiceWatchTimeout - not something
// normal operation should ever hit.
const enableStatefullWatchTimeout = 60 * time.Second

func (impl *Impl) EnablePlugin(ctx context.Context, req *pb.EnablePlugin_Request) (
	*pb.EnablePlugin_Response, error) {

	var err error
	switch req.GetPlugin() {
	case pb.VervPluginType_statefull_pg:
		payload, ok := req.Payload.(*pb.EnablePlugin_Request_StatefullCluster)
		if !ok {
			return nil, rerrors.New("invalid payload", codes.InvalidArgument)
		}

		// state.PgName is the fixed container name the postgres statefull
		// cluster is always deployed under - there's exactly one per node,
		// so it doubles as the task's entityID (mirrors CreateSmerd/
		// AssembleConfig/ConnectService keying tasks off their own natural
		// per-request identifier).
		initialContext := &pb.EnableStatefullTaskPayload{
			Request: payload.StatefullCluster,
		}

		_, err = impl.jobsEngine.Enqueue(ctx, state.PgName, jobs.EnableStatefullAction, initialContext)
		if err == nil {
			watchCtx, cancel := context.WithTimeout(ctx, enableStatefullWatchTimeout)
			defer cancel()

			var finalTask tasks_queries.VelezTask
			for task := range impl.jobsEngine.Watch(watchCtx, state.PgName, jobs.EnableStatefullAction) {
				finalTask = task
			}

			isDone := finalTask.Status == tasks_queries.VelezTaskStatusDONE
			isFailed := finalTask.Status == tasks_queries.VelezTaskStatusFAILED

			if !isDone && !isFailed && watchCtx.Err() != nil {
				err = rerrors.Wrapf(
					watchCtx.Err(),
					"timed out waiting for enable_statefull_mode task, last status: %q",
					finalTask.Status,
				)
			} else if finalTask.Status == tasks_queries.VelezTaskStatusFAILED {
				err = rerrors.New(finalTask.Error.String)
			}
		}
	default:
		return nil, rerrors.Wrap(errUnsupportedService)
	}

	if err != nil {
		return &pb.EnablePlugin_Response{}, rerrors.Wrap(err)
	}

	return &pb.EnablePlugin_Response{}, nil
}
