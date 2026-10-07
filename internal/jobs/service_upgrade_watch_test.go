package jobs

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/sqlc-dev/pqtype"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	upgradeWatchServiceName = "svc"
)

// upgradeWatchStep is the world as one poll sees it.
type upgradeWatchStep struct {
	isScheduled bool
	latest      sql.Null[tasks_queries.VelezTask]
	jobs        []JobStatus
}

// fakeUpgradeWorld advances one step per deployments List call - exactly one per poll - and stays on the last step.
type fakeUpgradeWorld struct {
	storage.DeploymentsStorage

	steps []upgradeWatchStep
	index int
}

func (f *fakeUpgradeWorld) Deployments() storage.DeploymentsStorage { return f }

func (f *fakeUpgradeWorld) List(_ context.Context, _ domain.ListDeploymentsReq) ([]domain.Deployment, error) {
	f.index = min(f.index+1, len(f.steps))

	if !f.current().isScheduled {
		return nil, nil
	}

	return []domain.Deployment{{Id: 1, SpecId: 5}}, nil
}

func (f *fakeUpgradeWorld) GetSpecificationById(
	_ context.Context, _ int64,
) (deployments_queries.GetSpecificationByIdRow, error) {
	row := deployments_queries.GetSpecificationByIdRow{
		VervPayload: pqtype.NullRawMessage{
			RawMessage: []byte(`{"name":"svc","environment":"default"}`),
			Valid:      true,
		},
	}

	return row, nil
}

func (f *fakeUpgradeWorld) Latest(
	_ context.Context, entityID, action string,
) (sql.Null[tasks_queries.VelezTask], error) {
	if entityID != SmerdEntityID("default", upgradeWatchServiceName) || action != UpgradeSmerdAction {
		return sql.Null[tasks_queries.VelezTask]{}, nil
	}

	return f.current().latest, nil
}

func (f *fakeUpgradeWorld) ListJobs(_ context.Context, _ tasks_queries.VelezTask) ([]JobStatus, error) {
	return f.current().jobs, nil
}

func (f *fakeUpgradeWorld) current() upgradeWatchStep {
	return f.steps[f.index-1]
}

func upgradeTask(id int64, status tasks_queries.VelezTaskStatus) sql.Null[tasks_queries.VelezTask] {
	return sql.Null[tasks_queries.VelezTask]{
		V:     tasks_queries.VelezTask{ID: id, Status: status},
		Valid: true,
	}
}

func upgradeJobs(status jobs_queries.VelezJobStatus) []JobStatus {
	return []JobStatus{{Name: "pull", Status: status}, {Name: "recreate"}}
}

func collectUpgradeSnapshots(t *testing.T, steps []upgradeWatchStep) []ServiceUpgradeSnapshot {
	t.Helper()

	world := &fakeUpgradeWorld{steps: steps}
	watcher := &ServiceUpgradeWatcher{
		deployments:  world,
		tasks:        world,
		pollInterval: time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	var got []ServiceUpgradeSnapshot

	for snapshot := range watcher.Watch(ctx, upgradeWatchServiceName) {
		got = append(got, snapshot)
	}

	require.NoError(t, ctx.Err(), "watcher did not close the stream")

	return got
}

func Test_ServiceUpgradeWatcher_Sequence(t *testing.T) {
	t.Parallel()

	pending := tasks_queries.VelezTaskStatusPENDING
	running := tasks_queries.VelezTaskStatusRUNNING
	done := tasks_queries.VelezTaskStatusDONE
	failed := tasks_queries.VelezTaskStatusFAILED

	type wanted struct {
		isTask bool
		id     int64
		status tasks_queries.VelezTaskStatus
	}

	cases := []struct {
		name  string
		steps []upgradeWatchStep
		want  []wanted
	}{
		{
			name:  "nothing in flight closes without a snapshot",
			steps: []upgradeWatchStep{{}},
			want:  nil,
		},
		{
			name: "scheduled then running then row gone ends with the done task",
			steps: []upgradeWatchStep{
				{isScheduled: true},
				{isScheduled: true},
				{isScheduled: true, latest: upgradeTask(1, pending), jobs: upgradeJobs("")},
				{isScheduled: true, latest: upgradeTask(1, running), jobs: upgradeJobs(jobs_queries.VelezJobStatusRUNNING)},
				{isScheduled: true, latest: upgradeTask(1, running), jobs: upgradeJobs(jobs_queries.VelezJobStatusRUNNING)},
				{latest: upgradeTask(1, done), jobs: upgradeJobs(jobs_queries.VelezJobStatusDONE)},
			},
			want: []wanted{{}, {true, 1, pending}, {true, 1, running}, {true, 1, done}},
		},
		{
			name: "task finished before the row flipped ends the stream",
			steps: []upgradeWatchStep{
				{isScheduled: true},
				{isScheduled: true, latest: upgradeTask(1, running)},
				{isScheduled: true, latest: upgradeTask(1, done)},
			},
			want: []wanted{{}, {true, 1, running}, {true, 1, done}},
		},
		{
			name: "an older finished task is ignored while the new upgrade is queued",
			steps: []upgradeWatchStep{
				{isScheduled: true, latest: upgradeTask(1, done)},
				{isScheduled: true, latest: upgradeTask(1, done)},
				{isScheduled: true, latest: upgradeTask(2, running)},
				{latest: upgradeTask(2, done)},
			},
			want: []wanted{{}, {true, 2, running}, {true, 2, done}},
		},
		{
			name: "a failed task surfaces as the terminal snapshot",
			steps: []upgradeWatchStep{
				{isScheduled: true},
				{isScheduled: true, latest: upgradeTask(3, running)},
				{latest: upgradeTask(3, failed)},
			},
			want: []wanted{{}, {true, 3, running}, {true, 3, failed}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := collectUpgradeSnapshots(t, tc.steps)

			require.Len(t, got, len(tc.want))

			for i, want := range tc.want {
				require.Equal(t, want.isTask, got[i].Task.Valid, "snapshot %d", i)

				if want.isTask {
					require.Equal(t, want.id, got[i].Task.V.ID, "snapshot %d", i)
					require.Equal(t, want.status, got[i].Task.V.Status, "snapshot %d", i)
				}
			}
		})
	}
}

func Test_ServiceUpgradeWatcher_RunningSnapshotCarriesJobs(t *testing.T) {
	t.Parallel()

	steps := []upgradeWatchStep{
		{isScheduled: true},
		{
			isScheduled: true,
			latest:      upgradeTask(1, tasks_queries.VelezTaskStatusRUNNING),
			jobs:        upgradeJobs(jobs_queries.VelezJobStatusRUNNING),
		},
		{latest: upgradeTask(1, tasks_queries.VelezTaskStatusDONE), jobs: upgradeJobs(jobs_queries.VelezJobStatusDONE)},
	}

	got := collectUpgradeSnapshots(t, steps)

	require.Len(t, got, 3)
	require.Nil(t, got[0].Jobs)
	require.Equal(t, upgradeJobs(jobs_queries.VelezJobStatusRUNNING), got[1].Jobs)
	require.Equal(t, upgradeJobs(jobs_queries.VelezJobStatusDONE), got[2].Jobs)
}
