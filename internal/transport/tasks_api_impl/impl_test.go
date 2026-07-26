package tasks_api_impl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"google.golang.org/grpc/metadata"
)

// fakeJobsEngine is a minimal hand-written fake for jobs.Engine, following
// this repo's convention (see internal/jobs/fakes_test.go) of small
// hand-written fakes over generated mocks for narrow interfaces. Enqueue
// just records what it was called with and returns a canned response; Watch
// replays a preset sequence of tasks_queries.VelezTask values onto a channel,
// mirroring the real engine.Watch's contract of closing the channel once the
// sequence is exhausted (or ctx is cancelled).
type fakeJobsEngine struct {
	enqueueResp tasks_queries.VelezTask
	enqueueErr  error

	enqueueCalls []enqueueCall
	watchCalls   []watchCall

	watchSequence []tasks_queries.VelezTask
}

type enqueueCall struct {
	entityID string
	action   string
}

type watchCall struct {
	entityID string
	action   string
}

func (f *fakeJobsEngine) Enqueue(_ context.Context, entityID, action string, _ any) (tasks_queries.VelezTask, error) {
	f.enqueueCalls = append(f.enqueueCalls, enqueueCall{entityID: entityID, action: action})

	return f.enqueueResp, f.enqueueErr
}

func (f *fakeJobsEngine) Watch(ctx context.Context, entityID, action string) <-chan tasks_queries.VelezTask {
	f.watchCalls = append(f.watchCalls, watchCall{entityID: entityID, action: action})

	ch := make(chan tasks_queries.VelezTask)

	go func() {
		defer close(ch)

		for _, task := range f.watchSequence {
			select {
			case ch <- task:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

// fakeTaskStatusStream is a minimal fake grpc.ServerStreamingServer[velez_api.TaskStatus]
// - only Send and Context are exercised by CreateSmerdStream, so the rest of
// the grpc.ServerStream surface is left as no-op stubs.
type fakeTaskStatusStream struct {
	ctx  context.Context
	sent []*velez_api.TaskStatus
}

func (f *fakeTaskStatusStream) Send(status *velez_api.TaskStatus) error {
	f.sent = append(f.sent, status)

	return nil
}

func (f *fakeTaskStatusStream) Context() context.Context { return f.ctx }

func (f *fakeTaskStatusStream) SetHeader(metadata.MD) error  { return nil }
func (f *fakeTaskStatusStream) SendHeader(metadata.MD) error { return nil }
func (f *fakeTaskStatusStream) SetTrailer(metadata.MD)       {}
func (f *fakeTaskStatusStream) SendMsg(any) error            { return nil }
func (f *fakeTaskStatusStream) RecvMsg(any) error            { return nil }

func Test_CreateSmerdStream(t *testing.T) {
	t.Parallel()

	now := time.Now()

	engine := &fakeJobsEngine{
		enqueueResp: tasks_queries.VelezTask{
			ID:       1,
			EntityID: "my-smerd",
			Action:   jobs.CreateSmerdAction,
			Status:   tasks_queries.VelezTaskStatusPENDING,
		},
		watchSequence: []tasks_queries.VelezTask{
			{ID: 1, EntityID: "my-smerd", Action: jobs.CreateSmerdAction, Status: tasks_queries.VelezTaskStatusPENDING, UpdatedAt: now},
			{ID: 1, EntityID: "my-smerd", Action: jobs.CreateSmerdAction, Status: tasks_queries.VelezTaskStatusRUNNING, UpdatedAt: now.Add(time.Second)},
			{ID: 1, EntityID: "my-smerd", Action: jobs.CreateSmerdAction, Status: tasks_queries.VelezTaskStatusDONE, UpdatedAt: now.Add(2 * time.Second)},
		},
	}

	impl := New(engine)

	req := &velez_api.CreateSmerd_Request{}
	req.Name = "my-smerd"

	stream := &fakeTaskStatusStream{ctx: context.Background()}

	err := impl.CreateSmerdStream(req, stream)
	require.NoError(t, err)

	// Enqueue and Watch must key off the same (entityID, action) pair the
	// unary CreateSmerd uses, so an in-flight unary create and a stream call
	// for the same name attach to the same task instead of double-creating.
	require.Len(t, engine.enqueueCalls, 1)
	require.Equal(t, "my-smerd", engine.enqueueCalls[0].entityID)
	require.Equal(t, jobs.CreateSmerdAction, engine.enqueueCalls[0].action)

	require.Len(t, engine.watchCalls, 1)
	require.Equal(t, "my-smerd", engine.watchCalls[0].entityID)
	require.Equal(t, jobs.CreateSmerdAction, engine.watchCalls[0].action)

	// The stream must have received every status update in order, and closed
	// cleanly (CreateSmerdStream returned nil) once the task reached DONE.
	require.Len(t, stream.sent, 3)
	require.Equal(t, velez_api.TaskStatus_PENDING, stream.sent[0].GetStatus())
	require.Equal(t, velez_api.TaskStatus_RUNNING, stream.sent[1].GetStatus())
	require.Equal(t, velez_api.TaskStatus_DONE, stream.sent[2].GetStatus())
}

func Test_CreateSmerdStream_EnqueueError(t *testing.T) {
	t.Parallel()

	engine := &fakeJobsEngine{
		enqueueErr: errors.New("boom"),
	}

	impl := New(engine)

	req := &velez_api.CreateSmerd_Request{}
	req.Name = "my-smerd"

	stream := &fakeTaskStatusStream{ctx: context.Background()}

	err := impl.CreateSmerdStream(req, stream)
	require.Error(t, err)
	require.Empty(t, stream.sent)
	// Watch must never be called when Enqueue fails.
	require.Empty(t, engine.watchCalls)
}
