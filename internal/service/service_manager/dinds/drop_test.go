package dinds

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testDindName      = "my-dind"
	testDindServiceId = 5
)

type testServices struct {
	storage.ServicesStorage
}

func (testServices) GetByName(_ context.Context, name string) (domain.Service, error) {
	svc := domain.Service{ID: testDindServiceId}

	svc.Name = name

	return svc, nil
}

type testDindInstances struct {
	storage.DindInstancesStorage
}

func (testDindInstances) GetDindInstanceByServiceId(_ context.Context, _ int64) (domain.DindInstance, error) {
	return domain.DindInstance{}, nil
}

type testRunners struct {
	storage.RunnersStorage

	runners []domain.Runner
}

func (r testRunners) ListRunners(_ context.Context) ([]domain.Runner, error) {
	return r.runners, nil
}

type testStorage struct {
	storage.Storage

	runners testRunners
}

func (testStorage) Services() storage.ServicesStorage { return testServices{} }

func (testStorage) DindInstances() storage.DindInstancesStorage { return testDindInstances{} }

func (s testStorage) Runners() storage.RunnersStorage { return s.runners }

type testEngine struct {
	jobs.Engine

	entityId string
	action   string
}

func (e *testEngine) EnqueueReplacing(
	_ context.Context, entityId, action string, payload any,
) (tasks_queries.VelezTask, error) {
	e.entityId = entityId
	e.action = action

	return tasks_queries.VelezTask{}, nil
}

func Test_DropDind_RejectsSynchronouslyWhenRunnerUsesDind(t *testing.T) {
	engine := &testEngine{}
	stg := testStorage{runners: testRunners{runners: []domain.Runner{{DindServiceId: testDindServiceId}}}}
	svc := New(stg, nil, engine, nil)

	err := svc.DropDind(t.Context(), testDindName)

	require.ErrorIs(t, err, user_errors.ErrDindInUse)
	require.Empty(t, engine.action)
}

func Test_DropDind_EnqueuesDropTaskWhenUnused(t *testing.T) {
	engine := &testEngine{}
	svc := New(testStorage{}, nil, engine, nil)

	err := svc.DropDind(t.Context(), testDindName)
	require.NoError(t, err)

	require.Equal(t, testDindName, engine.entityId)
	require.Equal(t, jobs.DropDindAction, engine.action)
}

func Test_CreateDind_RejectsInvalidInstanceNameBeforeAnyWork(t *testing.T) {
	engine := &testEngine{}
	svc := New(testStorage{}, nil, engine, nil)

	err := svc.CreateDind(t.Context(), domain.CreateDindReq{Name: "Bad Name"})

	require.ErrorIs(t, err, user_errors.ErrInvalidInstanceName)
	require.Empty(t, engine.action)
}
