package tasks_api_impl

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Impl struct {
	velez_api.UnimplementedTasksApiServer

	jobsEngine jobs.Engine
}

func New(jobsEngine jobs.Engine) *Impl {
	return &Impl{
		jobsEngine: jobsEngine,
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	velez_api.RegisterTasksApiServer(server, impl)
}

func (impl *Impl) Gateway(ctx context.Context, endpoint string, opts ...grpc.DialOption) (route string, handler http.Handler) {
	gwHttpMux := runtime.NewServeMux()

	err := velez_api.RegisterTasksApiHandlerFromEndpoint(
		ctx,
		gwHttpMux,
		endpoint,
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/tasks/", gwHttpMux
}

func (impl *Impl) WatchTask(req *velez_api.WatchTask_Request, stream grpc.ServerStreamingServer[velez_api.TaskStatus]) error {
	ctx := stream.Context()

	for task := range impl.jobsEngine.Watch(ctx, req.GetEntityId(), req.GetAction()) {
		err := stream.Send(taskToProto(task))
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateSmerdStream is an additive streaming pilot alongside VelezAPI's
// existing unary CreateSmerd: it enqueues the same create_smerd task
// (dedup'd on the same (name, action) key the unary RPC uses) and forwards
// TaskStatus updates as the task progresses, instead of blocking until it
// completes. The unary CreateSmerd stays untouched.
func (impl *Impl) CreateSmerdStream(req *velez_api.CreateSmerd_Request, stream grpc.ServerStreamingServer[velez_api.TaskStatus]) error {
	ctx := stream.Context()

	initialContext := &velez_api.CreateSmerdTaskPayload{}
	initialContext.SetRequest(req)

	_, err := impl.jobsEngine.Enqueue(ctx, req.GetName(), jobs.CreateSmerdAction, initialContext)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing create_smerd task")
	}

	for task := range impl.jobsEngine.Watch(ctx, req.GetName(), jobs.CreateSmerdAction) {
		err = stream.Send(taskToProto(task))
		if err != nil {
			return err
		}
	}

	return nil
}

func taskToProto(task tasks_queries.VelezTask) *velez_api.TaskStatus {
	out := &velez_api.TaskStatus{
		Status:    taskStatusToProto(task.Status),
		UpdatedAt: timestamppb.New(task.UpdatedAt),
	}

	if task.Error.Valid {
		out.Error = &task.Error.String
	}

	return out
}

func taskStatusToProto(status tasks_queries.VelezTaskStatus) velez_api.TaskStatus_Status {
	switch status {
	case tasks_queries.VelezTaskStatusPENDING:
		return velez_api.TaskStatus_PENDING
	case tasks_queries.VelezTaskStatusRUNNING:
		return velez_api.TaskStatus_RUNNING
	case tasks_queries.VelezTaskStatusDONE:
		return velez_api.TaskStatus_DONE
	case tasks_queries.VelezTaskStatusFAILED:
		return velez_api.TaskStatus_FAILED
	default:
		return velez_api.TaskStatus_UNKNOWN
	}
}
