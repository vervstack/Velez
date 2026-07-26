package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
	"github.com/sqlc-dev/pqtype"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	testVolumeName      = "myvol"
	testPathDataA       = "/data/a.txt"
	testPathDataB       = "/data/b.txt"
	testPathDataDir     = "/data"
	testLoaderContID    = "loader123"
	stepCreateContainer = "create_container"
)

func TestCopyToVolumeHandler_Action(t *testing.T) {
	h := NewCopyToVolumeHandler(nil)

	if h.Action() != CopyToVolumeAction {
		t.Errorf("expected action %q, got %q", CopyToVolumeAction, h.Action())
	}
}

func TestCopyToVolumeHandler_NewContext(t *testing.T) {
	h := NewCopyToVolumeHandler(nil)

	if _, ok := h.NewContext().(*velez_api.CopyToVolumeTaskPayload); !ok {
		t.Fatal("expected NewContext to return *velez_api.CopyToVolumeTaskPayload")
	}
}

// sortedFilePaths / mountedFolders are pure helpers extracted specifically
// so the ordering guarantee BuildJobs depends on is unit-testable in
// isolation from Docker.

func TestSortedFilePaths(t *testing.T) {
	m := map[string][]byte{
		"/data/c.txt": []byte("c"),
		testPathDataA: []byte("a"),
		testPathDataB: []byte("b"),
	}

	got := sortedFilePaths(m)
	want := []string{testPathDataA, testPathDataB, "/data/c.txt"}

	if !equalStrings(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestSortedFilePaths_Empty(t *testing.T) {
	got := sortedFilePaths(nil)
	if len(got) != 0 {
		t.Errorf("expected no keys, got %v", got)
	}
}

func TestMountedFolders_DedupPreservesFirstSeenOrder(t *testing.T) {
	sortedPaths := []string{"/data/a/1.txt", "/data/a/2.txt", "/data/b/1.txt"}

	got := mountedFolders(testVolumeName, sortedPaths)

	if len(got) != 2 {
		t.Fatalf("expected 2 distinct folders, got %d: %v", len(got), got)
	}

	if got[0].GetContainerPath() != "/data/a" || got[1].GetContainerPath() != "/data/b" {
		t.Errorf("expected folders [/data/a /data/b], got [%s %s]", got[0].GetContainerPath(), got[1].GetContainerPath())
	}

	if got[0].GetVolumeName() != testVolumeName || got[1].GetVolumeName() != testVolumeName {
		t.Errorf("expected volume name 'myvol' on every entry, got %v", got)
	}
}

func TestMountedFolders_Empty(t *testing.T) {
	got := mountedFolders(testVolumeName, nil)
	if len(got) != 0 {
		t.Errorf("expected no folders, got %v", got)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// TestCopyToVolumeHandler_BuildJobs_Deterministic is the migration's core
// regression test: BuildJobs runs fresh on every task run and every resume,
// and DONE jobs are skipped by (task_id, job_name). Go randomizes map
// iteration order, so without sorting, copy_file_<n> could map to a
// different file on every call and corrupt checkpoints. Calling BuildJobs
// twice on the same payload must produce identical job names, in identical
// order, every time.
func TestCopyToVolumeHandler_BuildJobs_Deterministic(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{
		VolumeName: testVolumeName,
		PathToFiles: map[string][]byte{
			"/data/zzz.txt":  []byte("z"),
			"/data/aaa.txt":  []byte("a"),
			"/data/mmm.txt":  []byte("m"),
			"/other/xxx.txt": []byte("x"),
		},
	}

	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)
	h := NewCopyToVolumeHandler(nodeClients)

	runs := make([][]string, 0, 10)

	for range 10 {
		namedJobs := h.BuildJobs(payload)

		names := make([]string, len(namedJobs))
		for i, nj := range namedJobs {
			names[i] = nj.Name
		}

		runs = append(runs, names)
	}

	first := runs[0]
	for i, names := range runs {
		if !equalStrings(names, first) {
			t.Fatalf("run %d produced job names %v, expected identical to run 0's %v", i, names, first)
		}
	}

	wantCount := 3 + 4 // create + start + drop + 4 files
	if len(first) != wantCount {
		t.Fatalf("expected %d jobs, got %d: %v", wantCount, len(first), first)
	}

	if first[0] != stepCreateContainer || first[1] != stepStartSidecar {
		t.Errorf("expected fixed jobs first, got %v", first[:2])
	}

	if first[len(first)-1] != stepDropContainer {
		t.Errorf("expected drop_container last, got %q", first[len(first)-1])
	}
}

// TestCopyToVolumeHandler_BuildJobs_NoFiles covers the N=0 edge case: only
// the 3 fixed jobs should be built.
func TestCopyToVolumeHandler_BuildJobs_NoFiles(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{
		VolumeName: testVolumeName,
	}

	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)
	h := NewCopyToVolumeHandler(nodeClients)

	namedJobs := h.BuildJobs(payload)
	if len(namedJobs) != 3 {
		t.Fatalf("expected 3 fixed jobs with no files, got %d", len(namedJobs))
	}

	wantNames := []string{stepCreateContainer, stepStartSidecar, stepDropContainer}
	for i, nj := range namedJobs {
		if nj.Name != wantNames[i] {
			t.Errorf("expected job %d named %q, got %q", i, wantNames[i], nj.Name)
		}
	}
}

// TestCopyToVolumeHandler_BuildJobs_CopyFileNamesMapToSortedPaths pins the
// exact copy_file_<n> -> file mapping so a future refactor can't silently
// break the ordering guarantee tested above.
func TestCopyToVolumeHandler_BuildJobs_CopyFileNamesMapToSortedPaths(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{
		VolumeName: testVolumeName,
		PathToFiles: map[string][]byte{
			testPathDataB: []byte("b"),
			testPathDataA: []byte("a"),
		},
	}

	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)
	h := NewCopyToVolumeHandler(nodeClients)

	namedJobs := h.BuildJobs(payload)

	job0, ok := namedJobs[2].Job.(*copyFileJob)
	if !ok {
		t.Fatalf("expected namedJobs[2] to be a *copyFileJob, got %T", namedJobs[2].Job)
	}

	if namedJobs[2].Name != "copy_file_0" || job0.filePath != testPathDataA {
		t.Errorf("expected copy_file_0 -> /data/a.txt, got %s -> %s", namedJobs[2].Name, job0.filePath)
	}

	job1, ok := namedJobs[3].Job.(*copyFileJob)
	if !ok {
		t.Fatalf("expected namedJobs[3] to be a *copyFileJob, got %T", namedJobs[3].Job)
	}

	if namedJobs[3].Name != "copy_file_1" || job1.filePath != testPathDataB {
		t.Errorf("expected copy_file_1 -> /data/b.txt, got %s -> %s", namedJobs[3].Name, job1.filePath)
	}
}

// createLoaderContainerJob depends only on the narrow node_clients.Docker
// interface (no ContainerInspect - see copy_to_volume.go's doc comment), so
// it's fully unit-testable with fakeDocker/fakeNodeClients.

func TestCreateLoaderContainerJob_Success(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{VolumeName: testVolumeName}

	docker := newFakeDocker()

	docker.containerCreateResp = container.CreateResponse{ID: testLoaderContID}

	nodeClients := newFakeNodeClients(docker)

	folders := mountedFolders(testVolumeName, []string{testPathDataA})

	j := &createLoaderContainerJob{nodeClients: nodeClients, req: payload, folders: folders, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetContainerId() != testLoaderContID {
		t.Errorf("expected container id 'loader123', got %q", payload.GetContainerId())
	}
}

// TestCreateLoaderContainerJob_ContainerCreateError is the required
// failure-path test: ContainerCreate's error must propagate, and no
// container id should be recorded.
func TestCreateLoaderContainerJob_ContainerCreateError(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{VolumeName: testVolumeName}

	docker := newFakeDocker()

	docker.containerCreateErr = errors.New("no space left on device")

	nodeClients := newFakeNodeClients(docker)

	j := &createLoaderContainerJob{nodeClients: nodeClients, req: payload, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when ContainerCreate fails")
	}

	if payload.GetContainerId() != "" {
		t.Errorf("expected no container id set on failure, got %q", payload.GetContainerId())
	}
}

func TestCreateLoaderContainerJob_Rollback_NoContainerId_NoOp(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}

	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)

	j := &createLoaderContainerJob{nodeClients: nodeClients, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docker.removeCalledWith) != 0 {
		t.Errorf("expected Remove not to be called, got %v", docker.removeCalledWith)
	}
}

func TestCreateLoaderContainerJob_Rollback_RemovesContainer(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}
	payload.SetContainerId(testLoaderContID)

	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)

	j := &createLoaderContainerJob{nodeClients: nodeClients, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docker.removeCalledWith) != 1 || docker.removeCalledWith[0] != testLoaderContID {
		t.Errorf("expected Remove called with 'loader123', got %v", docker.removeCalledWith)
	}
}

func TestCreateLoaderContainerJob_Rollback_NotFoundIsSwallowed(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}
	payload.SetContainerId(testLoaderContID)

	docker := newFakeDocker()

	docker.removeErr = errdefs.NotFound(errors.New("no such container"))

	nodeClients := newFakeNodeClients(docker)

	j := &createLoaderContainerJob{nodeClients: nodeClients, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("expected NotFound error to be swallowed, got: %v", err)
	}
}

// startLoaderContainerJob depends on the narrow startAPI interface -
// hand-fakeable with fakeContainerAPI without a full client.APIClient.

func TestStartLoaderContainerJob_Success(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}
	payload.SetContainerId(testLoaderContID)

	containerAPI := newFakeContainerAPI()

	j := &startLoaderContainerJob{dockerAPI: containerAPI, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(containerAPI.startCalledWith) != 1 || containerAPI.startCalledWith[0] != testLoaderContID {
		t.Errorf("expected ContainerStart called with 'loader123', got %v", containerAPI.startCalledWith)
	}
}

// TestStartLoaderContainerJob_NoContainerId_Error is a required failure-path
// test: Do must refuse to start without a container id.
func TestStartLoaderContainerJob_NoContainerId_Error(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}

	containerAPI := newFakeContainerAPI()

	j := &startLoaderContainerJob{dockerAPI: containerAPI, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error for an empty container id")
	}
}

// TestStartLoaderContainerJob_ContainerStartError is a required failure-path
// test: ContainerStart's error must propagate.
func TestStartLoaderContainerJob_ContainerStartError(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}
	payload.SetContainerId(testLoaderContID)

	containerAPI := newFakeContainerAPI()

	containerAPI.startErr = errors.New("docker daemon unreachable")

	j := &startLoaderContainerJob{dockerAPI: containerAPI, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when ContainerStart fails")
	}
}

func TestStartLoaderContainerJob_Rollback_StopsContainer(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}
	payload.SetContainerId(testLoaderContID)

	containerAPI := newFakeContainerAPI()

	j := &startLoaderContainerJob{dockerAPI: containerAPI, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(containerAPI.stopCalledWith) != 1 || containerAPI.stopCalledWith[0] != testLoaderContID {
		t.Errorf("expected ContainerStop called with 'loader123', got %v", containerAPI.stopCalledWith)
	}
}

func TestStartLoaderContainerJob_Rollback_NoContainerId_NoOp(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}

	containerAPI := newFakeContainerAPI()

	j := &startLoaderContainerJob{dockerAPI: containerAPI, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(containerAPI.stopCalledWith) != 0 {
		t.Errorf("expected ContainerStop not to be called, got %v", containerAPI.stopCalledWith)
	}
}

// copyFileJob combines mkdir (via node_clients.Docker.Exec) and the
// tar-based write (via the narrow copyAPI interface) into one checkpointed
// unit - see copy_to_volume.go's doc comment and questions.md #2.

func TestCopyFileJob_Success(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}
	payload.SetContainerId(testLoaderContID)

	docker := newFakeDocker()
	containerAPI := newFakeContainerAPI()

	j := &copyFileJob{
		docker:   docker,
		copyAPI:  containerAPI,
		ctx:      payload,
		filePath: testPathDataA,
		content:  []byte("hello"),
	}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docker.execCalledWith) != 1 {
		t.Fatalf("expected exactly 1 Exec call for mkdir, got %d", len(docker.execCalledWith))
	}

	wantCmd := []string{"mkdir", "-p", testPathDataDir}
	if !equalStrings(docker.execCalledWith[0].Cmd, wantCmd) {
		t.Errorf("expected mkdir command %v, got %v", wantCmd, docker.execCalledWith[0].Cmd)
	}

	if len(containerAPI.copyCalledWith) != 1 {
		t.Fatalf("expected exactly 1 CopyToContainer call, got %d", len(containerAPI.copyCalledWith))
	}

	call := containerAPI.copyCalledWith[0]
	if call.containerID != testLoaderContID {
		t.Errorf("expected copy to container 'loader123', got %q", call.containerID)
	}

	if call.dstPath != testPathDataDir {
		t.Errorf("expected copy dst path '/data', got %q", call.dstPath)
	}
}

// TestCopyFileJob_NilContent_Skipped preserves
// container_steps.CopyToContainer's existing behavior of silently skipping
// mount points whose content is nil (questions.md #3) - a present-but-nil
// []byte value in PathToFiles is possible since it's map[string][]byte.
func TestCopyFileJob_NilContent_Skipped(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}
	payload.SetContainerId(testLoaderContID)

	docker := newFakeDocker()
	containerAPI := newFakeContainerAPI()

	j := &copyFileJob{
		docker:   docker,
		copyAPI:  containerAPI,
		ctx:      payload,
		filePath: testPathDataA,
		content:  nil,
	}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docker.execCalledWith) != 0 {
		t.Errorf("expected no Exec call for nil content, got %d", len(docker.execCalledWith))
	}

	if len(containerAPI.copyCalledWith) != 0 {
		t.Errorf("expected no CopyToContainer call for nil content, got %d", len(containerAPI.copyCalledWith))
	}
}

// TestCopyFileJob_NoContainerId_Error is a required failure-path test.
func TestCopyFileJob_NoContainerId_Error(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}

	docker := newFakeDocker()
	containerAPI := newFakeContainerAPI()

	j := &copyFileJob{
		docker:   docker,
		copyAPI:  containerAPI,
		ctx:      payload,
		filePath: testPathDataA,
		content:  []byte("hello"),
	}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error for an empty container id")
	}
}

// TestCopyFileJob_ExecError is a required failure-path test: a failing
// mkdir must propagate and the write must never be attempted.
func TestCopyFileJob_ExecError(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}
	payload.SetContainerId(testLoaderContID)

	docker := newFakeDocker()

	docker.execErr = errors.New("exec failed: container not running")

	containerAPI := newFakeContainerAPI()

	j := &copyFileJob{
		docker:   docker,
		copyAPI:  containerAPI,
		ctx:      payload,
		filePath: testPathDataA,
		content:  []byte("hello"),
	}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when Exec (mkdir) fails")
	}

	if len(containerAPI.copyCalledWith) != 0 {
		t.Errorf("expected no CopyToContainer call after a failed mkdir, got %d", len(containerAPI.copyCalledWith))
	}
}

// TestCopyFileJob_CopyToContainerError is a required failure-path test.
func TestCopyFileJob_CopyToContainerError(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}
	payload.SetContainerId(testLoaderContID)

	docker := newFakeDocker()
	containerAPI := newFakeContainerAPI()

	containerAPI.copyErr = errors.New("no space left on device")

	j := &copyFileJob{
		docker:   docker,
		copyAPI:  containerAPI,
		ctx:      payload,
		filePath: testPathDataA,
		content:  []byte("hello"),
	}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when CopyToContainer fails")
	}
}

// dropLoaderContainerJob depends only on the narrow node_clients.Docker
// interface, so it's fully unit-testable with fakeDocker.

func TestDropLoaderContainerJob_NoContainerId_NoOp(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}

	docker := newFakeDocker()

	j := &dropLoaderContainerJob{docker: docker, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docker.removeCalledWith) != 0 {
		t.Errorf("expected Remove not to be called, got %v", docker.removeCalledWith)
	}
}

func TestDropLoaderContainerJob_Success(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}
	payload.SetContainerId(testLoaderContID)

	docker := newFakeDocker()

	j := &dropLoaderContainerJob{docker: docker, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docker.removeCalledWith) != 1 || docker.removeCalledWith[0] != testLoaderContID {
		t.Errorf("expected Remove called with 'loader123', got %v", docker.removeCalledWith)
	}
}

// TestDropLoaderContainerJob_RemoveErrorPropagates is a required
// failure-path test.
func TestDropLoaderContainerJob_RemoveErrorPropagates(t *testing.T) {
	payload := &velez_api.CopyToVolumeTaskPayload{}
	payload.SetContainerId(testLoaderContID)

	docker := newFakeDocker()

	docker.removeErr = errors.New("docker daemon unreachable")

	j := &dropLoaderContainerJob{docker: docker, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when Remove fails")
	}
}

// --- end-to-end coverage through taskWorker ---
//
// BuildJobs wires startLoaderContainerJob.dockerAPI and copyFileJob.copyAPI
// to nodeClients.Docker().Client(), the raw Docker engine SDK client
// (client.APIClient, 100+ methods). fakeDocker.Client() returns nil (see
// fakes_test.go / assemble_config_test.go's fetchConfigJob precedent) since
// hand-writing a full fake for that interface is out of scope. To still get
// true end-to-end coverage of BuildJobs' wiring plus the engine's
// checkpoint/rollback machinery, these tests build the job list via the real
// handler, then swap the two client.APIClient-typed fields for
// fakeContainerAPI (both are same-package unexported fields, and this test
// file is in package jobs) before handing the jobs to taskWorker.runJobs.

func copyToVolumeTask(
	t *testing.T, tasksStorage *fakeTasksStorage, entityID string, payload *velez_api.CopyToVolumeTaskPayload,
) tasks_queries.VelezTask {
	t.Helper()

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("unexpected error marshaling payload: %v", err)
	}

	taskCtx := pqtype.NullRawMessage{RawMessage: payloadJSON, Valid: true}
	params := tasks_queries.CreateTaskParams{
		EntityID: entityID,
		Action:   CopyToVolumeAction,
		Context:  taskCtx,
	}

	task, err := tasksStorage.CreateTask(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error creating task: %v", err)
	}

	return task
}

// patchClientAPIFields swaps every startLoaderContainerJob/copyFileJob's
// client.APIClient-typed field for containerAPI, in place.
func patchClientAPIFields(namedJobs []NamedJob, containerAPI *fakeContainerAPI) {
	for _, nj := range namedJobs {
		switch j := nj.Job.(type) {
		case *startLoaderContainerJob:
			j.dockerAPI = containerAPI
		case *copyFileJob:
			j.copyAPI = containerAPI
		}
	}
}

func TestCopyToVolumeHandler_HappyPath_EndToEnd(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()

	payload := &velez_api.CopyToVolumeTaskPayload{
		VolumeName: testVolumeName,
		PathToFiles: map[string][]byte{
			testPathDataA: []byte("A"),
			testPathDataB: []byte("B"),
		},
	}
	task := copyToVolumeTask(t, tasksStorage, testVolumeName, payload)

	docker := newFakeDocker()

	docker.containerCreateResp = container.CreateResponse{ID: testLoaderContID}

	nodeClients := newFakeNodeClients(docker)

	handler := NewCopyToVolumeHandler(nodeClients)

	taskCtx := handler.NewContext()

	err := json.Unmarshal(task.Context.RawMessage, taskCtx)
	if err != nil {
		t.Fatalf("unexpected error unmarshaling task context: %v", err)
	}

	namedJobs := handler.BuildJobs(taskCtx)

	containerAPI := newFakeContainerAPI()
	patchClientAPIFields(namedJobs, containerAPI)

	registry := NewRegistry()
	registry.Register(handler)

	w, ok := NewTaskWorker(tasksStorage, jobsStorage, registry, "test-worker", time.Hour).(*taskWorker)
	if !ok {
		t.Fatal("expected NewTaskWorker to return *taskWorker")
	}

	runErr := w.runJobs(context.Background(), task.ID, taskCtx, namedJobs)
	if runErr != nil {
		t.Fatalf("unexpected error running jobs: %v", runErr)
	}

	if len(containerAPI.startCalledWith) != 1 || containerAPI.startCalledWith[0] != testLoaderContID {
		t.Errorf("expected container started, got %v", containerAPI.startCalledWith)
	}

	if len(containerAPI.copyCalledWith) != 2 {
		t.Fatalf("expected 2 files copied, got %d", len(containerAPI.copyCalledWith))
	}

	firstDst := containerAPI.copyCalledWith[0].dstPath

	secondDst := containerAPI.copyCalledWith[1].dstPath
	if firstDst != testPathDataDir || secondDst != testPathDataDir {
		t.Errorf("expected both files copied under /data, got %v", containerAPI.copyCalledWith)
	}

	if len(docker.removeCalledWith) != 1 || docker.removeCalledWith[0] != testLoaderContID {
		t.Errorf("expected loader container dropped at the end, got %v", docker.removeCalledWith)
	}

	stepNames := []string{stepCreateContainer, stepStartSidecar, "copy_file_0", "copy_file_1", stepDropContainer}
	for _, name := range stepNames {
		row, ok := jobsStorage.rows[jobKey(task.ID, name)]
		if !ok {
			t.Errorf("expected a checkpoint row for job %q", name)

			continue
		}

		if row.Status != jobs_queries.VelezJobStatusDONE {
			t.Errorf("expected job %q checkpoint DONE, got %v", name, row.Status)
		}
	}
}

// TestCopyToVolumeHandler_FailurePath_CreateContainerFails is the required
// failure-path test for the first job in the chain: if ContainerCreate
// fails, the task must end up FAILED, no later job may run, and there's
// nothing to roll back (no container id was ever recorded).
func TestCopyToVolumeHandler_FailurePath_CreateContainerFails(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()

	payload := &velez_api.CopyToVolumeTaskPayload{
		VolumeName: testVolumeName,
		PathToFiles: map[string][]byte{
			testPathDataA: []byte("A"),
		},
	}
	task := copyToVolumeTask(t, tasksStorage, testVolumeName, payload)

	docker := newFakeDocker()

	docker.containerCreateErr = errors.New("no space left on device")

	nodeClients := newFakeNodeClients(docker)

	registry := NewRegistry()
	registry.Register(NewCopyToVolumeHandler(nodeClients))

	w, ok := NewTaskWorker(tasksStorage, jobsStorage, registry, "test-worker", time.Hour).(*taskWorker)
	if !ok {
		t.Fatal("expected NewTaskWorker to return *taskWorker")
	}

	err := w.run(context.Background(), task)
	if err == nil {
		t.Fatal("expected an error when the loader container fails to create")
	}

	finished := tasksStorage.get(task.ID)
	if finished.Status != tasks_queries.VelezTaskStatusFAILED {
		t.Errorf("expected task status FAILED, got %v", finished.Status)
	}

	if len(docker.removeCalledWith) != 0 {
		t.Errorf("expected Rollback to be a no-op (no container id was ever set), got Remove called with %v",
			docker.removeCalledWith)
	}
}

// TestCopyToVolumeHandler_FailurePath_LaterFileFailsCascadesRollback covers
// the checkpoint-correctness scenario the migration plan calls out
// explicitly: copy_file_1 fails after copy_file_0 already succeeded. The
// engine must roll back through start_container (stop) and create_container
// (remove), while copy_file_0's checkpoint stays DONE (jobs have no
// Rollback, matching the old pipeline's behavior of only ever undoing the
// container itself).
func TestCopyToVolumeHandler_FailurePath_LaterFileFailsCascadesRollback(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()

	payload := &velez_api.CopyToVolumeTaskPayload{
		VolumeName: testVolumeName,
		PathToFiles: map[string][]byte{
			testPathDataA: []byte("A"),
			testPathDataB: []byte("B"),
		},
	}
	task := copyToVolumeTask(t, tasksStorage, testVolumeName, payload)

	docker := newFakeDocker()

	docker.containerCreateResp = container.CreateResponse{ID: testLoaderContID}

	nodeClients := newFakeNodeClients(docker)

	handler := NewCopyToVolumeHandler(nodeClients)

	taskCtx := handler.NewContext()

	err := json.Unmarshal(task.Context.RawMessage, taskCtx)
	if err != nil {
		t.Fatalf("unexpected error unmarshaling task context: %v", err)
	}

	namedJobs := handler.BuildJobs(taskCtx)

	containerAPI := newFakeContainerAPI()

	containerAPI.copyErrOnCall = 2 // copy_file_0 (a.txt) succeeds, copy_file_1 (b.txt) fails.
	containerAPI.copyErr = errors.New("no space left on device")
	patchClientAPIFields(namedJobs, containerAPI)

	registry := NewRegistry()
	registry.Register(handler)

	w, ok := NewTaskWorker(tasksStorage, jobsStorage, registry, "test-worker", time.Hour).(*taskWorker)
	if !ok {
		t.Fatal("expected NewTaskWorker to return *taskWorker")
	}

	runErr := w.runJobs(context.Background(), task.ID, taskCtx, namedJobs)
	if runErr == nil {
		t.Fatal("expected an error when copy_file_1 fails")
	}

	if len(containerAPI.copyCalledWith) != 2 {
		t.Fatalf("expected both files attempted (0 succeeds, 1 fails), got %d calls", len(containerAPI.copyCalledWith))
	}

	if len(containerAPI.stopCalledWith) != 1 || containerAPI.stopCalledWith[0] != testLoaderContID {
		t.Errorf("expected rollback to stop the loader container, got %v", containerAPI.stopCalledWith)
	}

	if len(docker.removeCalledWith) != 1 || docker.removeCalledWith[0] != testLoaderContID {
		t.Errorf("expected rollback to remove the loader container, got %v", docker.removeCalledWith)
	}

	copyFile0Row, ok := jobsStorage.rows[jobKey(task.ID, "copy_file_0")]
	if !ok {
		t.Fatal("expected copy_file_0's checkpoint row to exist")
	}

	if copyFile0Row.Status != jobs_queries.VelezJobStatusDONE {
		t.Errorf("expected copy_file_0 checkpoint to stay DONE despite the later failure, got %v", copyFile0Row.Status)
	}
}
