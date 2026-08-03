package autoupgrade

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	testSmerdName     = "hello_world"
	testSmerdImage    = "godverv/hello_world:v0.0.15"
	testTaskFailedMsg = "upgrade blew up"
)

// enqueuedTask records one taskRunner.Enqueue call.
type enqueuedTask struct {
	entityID string
	action   string
	// contextJSON is the initial context marshaled exactly the way
	// jobs.Engine.Enqueue persists it.
	contextJSON []byte
}

// fakeTaskRunner is a pure in-process double for the jobs engine, recording
// what auto-upgrade enqueued and reporting a configurable terminal status.
// It fakes no external dependency - per CLAUDE.md this is the "records call
// order" category of test double.
type fakeTaskRunner struct {
	mu   sync.Mutex
	seen []enqueuedTask

	finalStatus tasks_queries.VelezTaskStatus
	finalError  sql.NullString
	// blockWatch makes Watch never emit, so the caller's watch timeout wins.
	blockWatch bool
}

func newFakeTaskRunner() *fakeTaskRunner {
	return &fakeTaskRunner{
		finalStatus: tasks_queries.VelezTaskStatusDONE,
	}
}

func (f *fakeTaskRunner) Enqueue(
	_ context.Context, entityID, action string, initialContext any,
) (tasks_queries.VelezTask, error) {
	contextJSON, err := json.Marshal(initialContext)
	if err != nil {
		return tasks_queries.VelezTask{}, rerrors.Wrap(err, "error marshaling initial context")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	task := enqueuedTask{
		entityID:    entityID,
		action:      action,
		contextJSON: contextJSON,
	}

	f.seen = append(f.seen, task)

	return tasks_queries.VelezTask{EntityID: entityID, Action: action}, nil
}

func (f *fakeTaskRunner) Watch(ctx context.Context, entityID, action string) <-chan tasks_queries.VelezTask {
	ch := make(chan tasks_queries.VelezTask, 1)

	if f.blockWatch {
		go func() {
			<-ctx.Done()
			close(ch)
		}()

		return ch
	}

	task := tasks_queries.VelezTask{
		EntityID: entityID,
		Action:   action,
		Status:   f.finalStatus,
		Error:    f.finalError,
	}

	ch <- task

	close(ch)

	return ch
}

func (f *fakeTaskRunner) calls() []enqueuedTask {
	f.mu.Lock()
	defer f.mu.Unlock()

	out := make([]enqueuedTask, len(f.seen))
	copy(out, f.seen)

	return out
}

func testUpgradeRequest() *velez_api.UpgradeSmerd_Request {
	return &velez_api.UpgradeSmerd_Request{
		Name:  testSmerdName,
		Image: testSmerdImage,
	}
}

// An auto-upgrade must become an upgrade_smerd task keyed on the smerd's
// name, replacing the deleted Pipeliner.UpgradeSmerd runner.
func TestAutoUpgrade_EnqueuesUpgradeSmerdTask(t *testing.T) {
	t.Parallel()

	runner := newFakeTaskRunner()
	au := &AutoUpgrade{jobsEngine: runner}

	err := au.upgrade(t.Context(), testUpgradeRequest())
	require.NoError(t, err)

	calls := runner.calls()
	require.Len(t, calls, 1)
	require.Equal(t, jobs.UpgradeSmerdAction, calls[0].action)
	require.Equal(t, testSmerdName, calls[0].entityID)

	payload := &velez_api.UpgradeSmerdTaskPayload{}

	err = json.Unmarshal(calls[0].contextJSON, payload)
	require.NoError(t, err)

	require.Equal(t, testSmerdName, payload.GetUpgradeRequest().GetName())
	require.Equal(t, testSmerdImage, payload.GetUpgradeRequest().GetImage())
}

// KNOWN LIMITATION pinned as a test: auto-upgrade is implicitly PROD-only.
// container.Summary carries no environment name, so the enqueued request's
// environment stays empty and resolves to the default environment - exactly
// what the pipeliner did. If this assertion ever starts failing, auto-upgrade
// has gained environment awareness and this comment (plus autoupgrade.go's)
// needs revisiting.
func TestAutoUpgrade_EnqueuedTaskIsProdOnly(t *testing.T) {
	t.Parallel()

	runner := newFakeTaskRunner()
	au := &AutoUpgrade{jobsEngine: runner}

	err := au.upgrade(t.Context(), testUpgradeRequest())
	require.NoError(t, err)

	payload := &velez_api.UpgradeSmerdTaskPayload{}

	err = json.Unmarshal(runner.calls()[0].contextJSON, payload)
	require.NoError(t, err)

	require.Empty(t, payload.GetUpgradeRequest().GetEnvironment())
}

// A FAILED task must surface as an error, the contract the pipeliner's
// runner.Run(ctx) error had - au.do's errgroup relies on it.
func TestAutoUpgrade_FailedTaskReturnsError(t *testing.T) {
	t.Parallel()

	runner := newFakeTaskRunner()

	runner.finalStatus = tasks_queries.VelezTaskStatusFAILED
	runner.finalError = sql.NullString{String: testTaskFailedMsg, Valid: true}

	au := &AutoUpgrade{jobsEngine: runner}

	err := au.upgrade(t.Context(), testUpgradeRequest())
	require.ErrorContains(t, err, testTaskFailedMsg)
}

// A task that never reaches a terminal status must time out rather than
// wedging the auto-upgrade loop forever.
func TestAutoUpgrade_TimesOutOnStuckTask(t *testing.T) {
	t.Parallel()

	runner := newFakeTaskRunner()

	runner.blockWatch = true

	au := &AutoUpgrade{jobsEngine: runner}

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	err := au.upgrade(ctx, testUpgradeRequest())
	require.ErrorContains(t, err, "timed out waiting for upgrade_smerd task")
}
