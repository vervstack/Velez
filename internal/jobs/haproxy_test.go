package jobs

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/local_storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	testHaproxyName = "database"
)

func TestHaproxyConfigRender(t *testing.T) {
	t.Parallel()

	const (
		master   = "pgaas_primary"
		user     = "primary_user"
		port     = 15432
		expected = `defaults
    mode tcp
    timeout connect 5s
    timeout client 30s
    timeout server 30s

listen pg
    bind *:15432
    option pgsql-check user primary_user
    server pgaas_primary pgaas_primary:5432 check
`
	)

	config, err := renderHaproxyConfig(port, master, user)
	require.NoError(t, err)
	require.Equal(t, expected, string(config))

	_, err = renderHaproxyConfig(port, "pgaas_primary\nserver other", user)
	require.Error(t, err)

	_, err = renderHaproxyConfig(port, master, "bad user\n")
	require.Error(t, err)

	_, err = renderHaproxyConfig(0, master, user)
	require.Error(t, err)

	_, err = renderHaproxyConfig(domain.HaproxyMaxPort+1, master, user)
	require.Error(t, err)
}

func TestHaproxyDescriptorUsesCustomPortAndVolume(t *testing.T) {
	t.Parallel()

	const port = 15432

	request := &velez_api.CreateHaproxyInstance_Request{Name: testHaproxyName}
	descriptor, err := haproxyDescriptor(request, port)
	require.NoError(t, err)

	app := descriptor.Deployment.App
	require.Equal(t, "haproxy:lts", app.Image)
	require.Len(t, app.Ports, 1)
	require.Equal(t, port, app.Ports[0].Port)
	require.Equal(t, port, app.Ports[0].ExposeTo)
	require.Len(t, app.Volumes, 1)
	require.Equal(t, domain.HaproxyConfigVolumeName(request.GetName()), app.Volumes[0].Name)
}

func TestHaproxySelectMasterSkipsCompletedSteps(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	taskStorage := local_storage.NewTasksStorage()
	jobStorage := local_storage.NewJobsStorage()

	createTask := tasks_queries.CreateTaskParams{EntityID: testHaproxyName, Action: SelectHaproxyMasterAction}
	task, err := taskStorage.CreateTask(ctx, createTask)
	require.NoError(t, err)

	request := &velez_api.SelectHaproxyMaster_Request{Name: testHaproxyName, MasterPgInstance: "pgaas_primary"}
	payload := &velez_api.SelectHaproxyMasterTaskPayload{Request: request}
	handler := NewSelectHaproxyMasterHandler(nil, nil, nil)
	steps := handler.BuildJobs(payload)

	require.Len(t, steps, 5)
	require.Equal(t, stepValidateHaproxyMaster, steps[0].Name)
	require.Equal(t, stepStageHaproxyConfig, steps[1].Name)
	require.Equal(t, stepCheckHaproxyConfig, steps[2].Name)
	require.Equal(t, stepReloadHaproxy, steps[3].Name)
	require.Equal(t, stepAwaitHaproxyStable, steps[4].Name)

	// None of these steps can run without its dependencies. A resumed task must
	// skip them once checkpointed, especially after its reload was signalled.
	for _, step := range steps {
		createJob := jobs_queries.CreateRunningJobParams{TaskID: task.ID, JobName: step.Name}

		_, err = jobStorage.CreateRunningJob(ctx, createJob)
		require.NoError(t, err)

		finishJob := jobs_queries.FinishJobParams{
			TaskID: task.ID, JobName: step.Name, Status: jobs_queries.VelezJobStatusDONE,
		}

		err = jobStorage.FinishJob(ctx, finishJob)
		require.NoError(t, err)

		job := Checkpoint(jobStorage, taskStorage, task.ID, step.Name, payload, step.Job)

		err = job.Do(ctx)
		require.NoError(t, err)
	}
}

func TestHaproxyStabilityCheck(t *testing.T) {
	t.Parallel()

	newInspect := func(state *container.State, restartCount int) container.InspectResponse {
		base := &container.ContainerJSONBase{State: state, RestartCount: restartCount}

		return container.InspectResponse{ContainerJSONBase: base}
	}

	tests := []struct {
		name       string
		info       container.InspectResponse
		wantStable bool
		wantReason string
	}{
		{
			name:       "running cleanly is stable",
			info:       newInspect(&container.State{Running: true}, 0),
			wantStable: true,
		},
		{
			name:       "missing state is a failure",
			info:       container.InspectResponse{},
			wantReason: "container state is unknown",
		},
		{
			name:       "restarting is a failure",
			info:       newInspect(&container.State{Running: true, Restarting: true}, 1),
			wantReason: "container is restarting (restart count 1)",
		},
		{
			name:       "non-zero restart count is a failure even while running",
			info:       newInspect(&container.State{Running: true}, 3),
			wantReason: "container restarted 3 time(s)",
		},
		{
			name:       "exited is a failure with the exit code",
			info:       newInspect(&container.State{Running: false, ExitCode: 1}, 0),
			wantReason: "container exited with code 1",
		},
		{
			name: "exited with a daemon error includes it",
			info: newInspect(
				&container.State{Running: false, ExitCode: 1, Error: "port already in use"}, 0),
			wantReason: "container exited with code 1: port already in use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			isStable, reason := haproxyStabilityCheck(tt.info)
			require.Equal(t, tt.wantStable, isStable)
			require.Equal(t, tt.wantReason, reason)
		})
	}
}
