package control_plane_api_impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/state"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

// fakeJobsEngine is a hand-written fake for jobs.Engine (no mocking library
// is used in this repo - see internal/jobs/fakes_test.go for the same
// pattern applied to other interfaces). It records whether Watch was called
// so tests can assert EnablePlugin no longer blocks on it.
type fakeJobsEngine struct {
	enqueueTask tasks_queries.VelezTask
	enqueueErr  error

	watchCalled bool
}

func (f *fakeJobsEngine) Enqueue(
	_ context.Context, _, _ string, _ any,
) (tasks_queries.VelezTask, error) {
	return f.enqueueTask, f.enqueueErr
}

func (f *fakeJobsEngine) Watch(_ context.Context, _, _ string) <-chan tasks_queries.VelezTask {
	f.watchCalled = true

	ch := make(chan tasks_queries.VelezTask)
	close(ch)

	return ch
}

func (f *fakeJobsEngine) ListJobs(_ context.Context, _ tasks_queries.VelezTask) ([]jobs.JobStatus, error) {
	return nil, nil
}

func (f *fakeJobsEngine) SetRegistry(_ *jobs.Registry) {}

func Test_EnablePlugin_StatefullPg_Success(t *testing.T) {
	fakeEngine := &fakeJobsEngine{
		enqueueTask: tasks_queries.VelezTask{
			EntityID: state.PgName(""),
			Action:   jobs.EnableStatefullAction,
		},
	}

	impl := &Impl{
		jobsEngine: fakeEngine,
	}

	statefullReq := &pb.EnableStatefullCluster{}
	payload := &pb.EnablePlugin_Request_StatefullCluster{
		StatefullCluster: statefullReq,
	}
	req := &pb.EnablePlugin_Request{
		Plugin:  pb.VervPluginType_statefull_pg,
		Payload: payload,
	}

	resp, err := impl.EnablePlugin(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, state.PgName(""), resp.GetEntityId())
	assert.Equal(t, jobs.EnableStatefullAction, resp.GetAction())
	assert.False(t, fakeEngine.watchCalled, "EnablePlugin must not block on Watch anymore")
}

func Test_EnablePlugin_StatefullPg_EnqueueError(t *testing.T) {
	fakeEngine := &fakeJobsEngine{
		enqueueErr: rerrors.New("enqueue boom"),
	}

	impl := &Impl{
		jobsEngine: fakeEngine,
	}

	statefullReq := &pb.EnableStatefullCluster{}
	payload := &pb.EnablePlugin_Request_StatefullCluster{
		StatefullCluster: statefullReq,
	}
	req := &pb.EnablePlugin_Request{
		Plugin:  pb.VervPluginType_statefull_pg,
		Payload: payload,
	}

	_, err := impl.EnablePlugin(context.Background(), req)
	require.Error(t, err)
	assert.False(t, fakeEngine.watchCalled)
}
