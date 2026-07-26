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
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/headscale"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	testServiceName  = "my_service"
	testIssuedKey    = "issued-key"
	testSvc          = "svc"
	testTailscaleImg = "tailscale/tailscale:v1.90.8"
	testSidecarID    = "sidecar123"
	testSvcTsSidecar = "svc-ts-sidecar"
)

func TestConnectServiceToVpnHandler_Action(t *testing.T) {
	h := NewConnectServiceToVpnHandler(nil, nil, nil)

	if h.Action() != ConnectServiceToVpnAction {
		t.Errorf("expected action %q, got %q", ConnectServiceToVpnAction, h.Action())
	}
}

func TestConnectServiceToVpnHandler_NewContext(t *testing.T) {
	h := NewConnectServiceToVpnHandler(nil, nil, nil)

	if _, ok := h.NewContext().(*velez_api.ConnectServiceToVpnTaskPayload); !ok {
		t.Fatal("expected NewContext to return *velez_api.ConnectServiceToVpnTaskPayload")
	}
}

func TestConnectServiceToVpnHandler_BuildJobs_NamesAndOrder(t *testing.T) {
	payload := &velez_api.ConnectServiceToVpnTaskPayload{ServiceName: testServiceName}

	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)
	h := NewConnectServiceToVpnHandler(nodeClients, newFakeVpnClient(), newFakeServiceDiscovery())

	namedJobs := h.BuildJobs(payload)

	wantNames := []string{
		stepCheckSidecar, stepPrepareNamespace, stepGetClientKey, stepGetLoginServerURL,
		stepPrepareSidecarImage, stepCreatePgContainer, stepStartSidecar, stepAddMakoshRecord,
	}
	if len(namedJobs) != len(wantNames) {
		t.Fatalf("expected %d jobs, got %d", len(wantNames), len(namedJobs))
	}

	for i, name := range wantNames {
		if namedJobs[i].Name != name {
			t.Errorf("expected job %d named %q, got %q", i, name, namedJobs[i].Name)
		}
	}
}

// prepareNamespaceJob

func TestPrepareNamespaceJob_ExistingNamespace(t *testing.T) {
	vpn := newFakeVpnClient()
	vpn.namespaces[testSvc] = domain.VcnNamespace{Id: "ns-existing", Name: testSvc}

	payload := &velez_api.ConnectServiceToVpnTaskPayload{ServiceName: testSvc}
	j := &prepareNamespaceJob{vpnClient: vpn, req: payload, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetNamespaceId() != "ns-existing" {
		t.Errorf("expected namespace id 'ns-existing', got %q", payload.GetNamespaceId())
	}
}

func TestPrepareNamespaceJob_CreatesWhenMissing(t *testing.T) {
	vpn := newFakeVpnClient()

	payload := &velez_api.ConnectServiceToVpnTaskPayload{ServiceName: testSvc}
	j := &prepareNamespaceJob{vpnClient: vpn, req: payload, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetNamespaceId() != "ns-svc" {
		t.Errorf("expected created namespace id 'ns-svc', got %q", payload.GetNamespaceId())
	}
}

func TestPrepareNamespaceJob_GetNamespaceError(t *testing.T) {
	vpn := newFakeVpnClient()
	vpn.getNamespaceErr = errors.New("headscale unreachable")

	payload := &velez_api.ConnectServiceToVpnTaskPayload{ServiceName: testSvc}
	j := &prepareNamespaceJob{vpnClient: vpn, req: payload, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when GetNamespace fails")
	}

	if payload.GetNamespaceId() != "" {
		t.Errorf("expected no namespace id set on failure, got %q", payload.GetNamespaceId())
	}
}

func TestPrepareNamespaceJob_CreateNamespaceError(t *testing.T) {
	vpn := newFakeVpnClient()
	vpn.createNamespaceErr = errors.New("quota exceeded")

	payload := &velez_api.ConnectServiceToVpnTaskPayload{ServiceName: testSvc}
	j := &prepareNamespaceJob{vpnClient: vpn, req: payload, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when CreateNamespace fails")
	}
}

// getClientKeyJob

func TestGetClientKeyJob_ExistingKeyFound(t *testing.T) {
	vpn := newFakeVpnClient()
	vpn.authKeyResp = domain.VcnAuthKey{Key: "existing-key"}

	payload := &velez_api.ConnectServiceToVpnTaskPayload{NamespaceId: proto("ns-1")}
	j := &getClientKeyJob{vpnClient: vpn, namespace: payload, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// See the Do() doc comment: unlike the original pipeline step, the found
	// key must actually be returned - the pointer-reassignment bug is
	// structurally impossible with a Set-based accessor.
	if payload.GetClientKey() != "existing-key" {
		t.Errorf("expected existing key 'existing-key' to be reused, got %q", payload.GetClientKey())
	}

	if vpn.issueCalls != 0 {
		t.Errorf("expected no new key to be issued when one already exists, got %d issue calls", vpn.issueCalls)
	}
}

func TestGetClientKeyJob_NotFoundIssuesNewKey(t *testing.T) {
	vpn := newFakeVpnClient()
	vpn.authKeyErr = headscale.ErrNotFound
	vpn.issuedKey = "brand-new-key"

	payload := &velez_api.ConnectServiceToVpnTaskPayload{NamespaceId: proto("ns-1")}
	j := &getClientKeyJob{vpnClient: vpn, namespace: payload, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetClientKey() != "brand-new-key" {
		t.Errorf("expected issued key 'brand-new-key', got %q", payload.GetClientKey())
	}

	if vpn.issueCalls != 1 {
		t.Errorf("expected exactly 1 issue call, got %d", vpn.issueCalls)
	}
}

func TestGetClientKeyJob_UnexpectedGetAuthKeyErrorPropagates(t *testing.T) {
	vpn := newFakeVpnClient()
	vpn.authKeyErr = errors.New("headscale down")

	payload := &velez_api.ConnectServiceToVpnTaskPayload{NamespaceId: proto("ns-1")}
	j := &getClientKeyJob{vpnClient: vpn, namespace: payload, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error to propagate when GetClientAuthKey fails with a non-NotFound error")
	}
}

func TestGetClientKeyJob_IssueClientKeyError(t *testing.T) {
	vpn := newFakeVpnClient()
	vpn.authKeyErr = headscale.ErrNotFound
	vpn.issueKeyErr = errors.New("headscale rejected the request")

	payload := &velez_api.ConnectServiceToVpnTaskPayload{NamespaceId: proto("ns-1")}
	j := &getClientKeyJob{vpnClient: vpn, namespace: payload, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when IssueClientKey fails")
	}
}

// getLoginServerURLJob

func TestGetLoginServerUrlJob_SetsConstant(t *testing.T) {
	payload := &velez_api.ConnectServiceToVpnTaskPayload{}
	j := &getLoginServerURLJob{ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetLoginServerUrl() != "https://vcn.redsock.ru" {
		t.Errorf("expected login server url constant, got %q", payload.GetLoginServerUrl())
	}
}

// prepareSidecarImageJob

func TestPrepareSidecarImageJob_Success(t *testing.T) {
	docker := newFakeDocker()
	req := &container.CreateRequest{Config: &container.Config{Image: testTailscaleImg}}

	j := &prepareSidecarImageJob{docker: docker, req: req}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrepareSidecarImageJob_PullImageError(t *testing.T) {
	docker := newFakeDocker()
	docker.pullImageErr = errors.New("registry unreachable")
	req := &container.CreateRequest{Config: &container.Config{Image: testTailscaleImg}}

	j := &prepareSidecarImageJob{docker: docker, req: req}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when PullImage fails")
	}
}

// createSidecarContainerJob

func TestCreateSidecarContainerJob_Success(t *testing.T) {
	docker := newFakeDocker()
	docker.containerCreateResp = container.CreateResponse{ID: testSidecarID}
	nodeClients := newFakeNodeClients(docker)

	launchContainer := &container.CreateRequest{Config: &container.Config{Image: testTailscaleImg}}
	payload := &velez_api.ConnectServiceToVpnTaskPayload{
		ClientKey:      proto("the-key"),
		LoginServerUrl: proto("https://vcn.redsock.ru"),
	}

	j := &createSidecarContainerJob{
		nodeClients:     nodeClients,
		launchContainer: launchContainer,
		containerName:   testSvcTsSidecar,
		hostname:        testSvcTsSidecar,
		clientKey:       payload,
		loginURL:        payload,
		ctx:             payload,
	}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetContainerId() != testSidecarID {
		t.Errorf("expected container id 'sidecar123', got %q", payload.GetContainerId())
	}

	wantEnv := []string{
		"TS_HOSTNAME=svc-ts-sidecar",
		"TS_AUTHKEY=the-key",
		"TS_EXTRA_ARGS=--login-server=https://vcn.redsock.ru",
	}

	gotEnv := launchContainer.Env
	if len(gotEnv) != len(wantEnv) {
		t.Fatalf("expected env %v, got %v", wantEnv, gotEnv)
	}

	for i := range wantEnv {
		if gotEnv[i] != wantEnv[i] {
			t.Errorf("expected env[%d]=%q, got %q", i, wantEnv[i], gotEnv[i])
		}
	}
}

func TestCreateSidecarContainerJob_ContainerCreateError(t *testing.T) {
	docker := newFakeDocker()
	docker.containerCreateErr = errors.New("no space left on device")
	nodeClients := newFakeNodeClients(docker)

	launchContainer := &container.CreateRequest{Config: &container.Config{Image: testTailscaleImg}}
	payload := &velez_api.ConnectServiceToVpnTaskPayload{}

	j := &createSidecarContainerJob{
		nodeClients:     nodeClients,
		launchContainer: launchContainer,
		containerName:   testSvcTsSidecar,
		hostname:        testSvcTsSidecar,
		clientKey:       payload,
		loginURL:        payload,
		ctx:             payload,
	}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when ContainerCreate fails")
	}

	if payload.GetContainerId() != "" {
		t.Errorf("expected no container id set on failure, got %q", payload.GetContainerId())
	}
}

func TestCreateSidecarContainerJob_Rollback_NoContainerId_NoOp(t *testing.T) {
	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)
	payload := &velez_api.ConnectServiceToVpnTaskPayload{}

	j := &createSidecarContainerJob{nodeClients: nodeClients, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docker.removeCalledWith) != 0 {
		t.Errorf("expected Remove not to be called, got %v", docker.removeCalledWith)
	}
}

func TestCreateSidecarContainerJob_Rollback_RemovesContainer(t *testing.T) {
	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)
	payload := &velez_api.ConnectServiceToVpnTaskPayload{ContainerId: proto(testSidecarID)}

	j := &createSidecarContainerJob{nodeClients: nodeClients, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docker.removeCalledWith) != 1 || docker.removeCalledWith[0] != testSidecarID {
		t.Errorf("expected Remove called with 'sidecar123', got %v", docker.removeCalledWith)
	}
}

func TestCreateSidecarContainerJob_Rollback_NotFoundIsSwallowed(t *testing.T) {
	docker := newFakeDocker()
	docker.removeErr = errdefs.NotFound(errors.New("no such container"))
	nodeClients := newFakeNodeClients(docker)
	payload := &velez_api.ConnectServiceToVpnTaskPayload{ContainerId: proto(testSidecarID)}

	j := &createSidecarContainerJob{nodeClients: nodeClients, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("expected a not-found Remove error to be swallowed, got %v", err)
	}
}

// startSidecarContainerJob

func TestStartSidecarContainerJob_Success(t *testing.T) {
	api := newFakeContainerAPI()
	payload := &velez_api.ConnectServiceToVpnTaskPayload{ContainerId: proto(testSidecarID)}

	j := &startSidecarContainerJob{dockerAPI: api, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(api.startCalledWith) != 1 || api.startCalledWith[0] != testSidecarID {
		t.Errorf("expected container started with 'sidecar123', got %v", api.startCalledWith)
	}
}

func TestStartSidecarContainerJob_NoContainerId_Error(t *testing.T) {
	api := newFakeContainerAPI()
	payload := &velez_api.ConnectServiceToVpnTaskPayload{}

	j := &startSidecarContainerJob{dockerAPI: api, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when no container id is set")
	}
}

func TestStartSidecarContainerJob_ContainerStartError(t *testing.T) {
	api := newFakeContainerAPI()
	api.startErr = errors.New("engine unavailable")
	payload := &velez_api.ConnectServiceToVpnTaskPayload{ContainerId: proto(testSidecarID)}

	j := &startSidecarContainerJob{dockerAPI: api, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when ContainerStart fails")
	}
}

func TestStartSidecarContainerJob_Rollback_StopsContainer(t *testing.T) {
	api := newFakeContainerAPI()
	payload := &velez_api.ConnectServiceToVpnTaskPayload{ContainerId: proto(testSidecarID)}

	j := &startSidecarContainerJob{dockerAPI: api, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(api.stopCalledWith) != 1 || api.stopCalledWith[0] != testSidecarID {
		t.Errorf("expected container stopped with 'sidecar123', got %v", api.stopCalledWith)
	}
}

// addMakoshRecordJob

func TestAddMakoshRecordJob_Success(t *testing.T) {
	sd := newFakeServiceDiscovery()

	j := &addMakoshRecordJob{sd: sd, serviceName: testSvc, hostname: testSvcTsSidecar}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sd.upsertCalledWith) != 1 {
		t.Fatalf("expected 1 upsert call, got %d", len(sd.upsertCalledWith))
	}

	endpoints := sd.upsertCalledWith[0].GetEndpoints()
	if len(endpoints) != 1 || endpoints[0].GetServiceName() != testSvc {
		t.Errorf("expected endpoint for service 'svc', got %v", endpoints)
	}

	if len(endpoints[0].GetAddrs()) != 1 || endpoints[0].GetAddrs()[0] != testSvcTsSidecar {
		t.Errorf("expected addr 'svc-ts-sidecar', got %v", endpoints[0].GetAddrs())
	}
}

func TestAddMakoshRecordJob_UpsertError(t *testing.T) {
	sd := newFakeServiceDiscovery()
	sd.upsertErr = errors.New("makosh unreachable")

	j := &addMakoshRecordJob{sd: sd, serviceName: testSvc, hostname: testSvcTsSidecar}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when UpsertEndpoints fails")
	}
}

// End-to-end coverage through the real handler + taskWorker, following
// copy_to_volume_test.go's patchClientAPIFields precedent: startSidecarContainerJob's
// dockerAPI field is nodeClients.Docker().Client() (nil on fakeDocker), so it's
// swapped for a fakeContainerAPI after BuildJobs before running.

func connectServiceToVpnTask(
	t *testing.T, tasksStorage *fakeTasksStorage, entityID string, payload *velez_api.ConnectServiceToVpnTaskPayload,
) tasks_queries.VelezTask {
	t.Helper()

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("unexpected error marshaling payload: %v", err)
	}

	taskCtx := pqtype.NullRawMessage{RawMessage: payloadJSON, Valid: true}
	params := tasks_queries.CreateTaskParams{
		EntityID: entityID,
		Action:   ConnectServiceToVpnAction,
		Context:  taskCtx,
	}

	task, err := tasksStorage.CreateTask(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error creating task: %v", err)
	}

	return task
}

func patchStartAPIField(namedJobs []NamedJob, containerAPI *fakeContainerAPI) {
	for _, nj := range namedJobs {
		if j, ok := nj.Job.(*startSidecarContainerJob); ok {
			j.dockerAPI = containerAPI
		}
	}
}

func TestConnectServiceToVpnHandler_HappyPath_EndToEnd(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()

	payload := &velez_api.ConnectServiceToVpnTaskPayload{ServiceName: testServiceName}
	task := connectServiceToVpnTask(t, tasksStorage, testServiceName, payload)

	docker := newFakeDocker()
	docker.containerCreateResp = container.CreateResponse{ID: testSidecarID}
	nodeClients := newFakeNodeClients(docker)

	vpn := newFakeVpnClient()
	vpn.issuedKey = testIssuedKey
	sd := newFakeServiceDiscovery()

	handler := NewConnectServiceToVpnHandler(nodeClients, vpn, sd)

	taskCtx := handler.NewContext()

	err := json.Unmarshal(task.Context.RawMessage, taskCtx)
	if err != nil {
		t.Fatalf("unexpected error unmarshaling task context: %v", err)
	}

	namedJobs := handler.BuildJobs(taskCtx)

	containerAPI := newFakeContainerAPI()
	patchStartAPIField(namedJobs, containerAPI)

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

	if len(containerAPI.startCalledWith) != 1 || containerAPI.startCalledWith[0] != testSidecarID {
		t.Errorf("expected container started, got %v", containerAPI.startCalledWith)
	}

	if len(sd.upsertCalledWith) != 1 {
		t.Errorf("expected 1 makosh upsert call, got %d", len(sd.upsertCalledWith))
	}

	finishedPayload, ok := taskCtx.(*velez_api.ConnectServiceToVpnTaskPayload)
	if !ok {
		t.Fatalf("expected taskCtx to be *velez_api.ConnectServiceToVpnTaskPayload, got %T", taskCtx)
	}

	if finishedPayload.GetNamespaceId() != "ns-my_service" {
		t.Errorf("expected namespace id 'ns-my_service', got %q", finishedPayload.GetNamespaceId())
	}

	if finishedPayload.GetClientKey() != testIssuedKey {
		t.Errorf("expected client key 'issued-key', got %q", finishedPayload.GetClientKey())
	}

	if finishedPayload.GetContainerId() != testSidecarID {
		t.Errorf("expected container id 'sidecar123', got %q", finishedPayload.GetContainerId())
	}

	for _, name := range []string{
		stepCheckSidecar, stepPrepareNamespace, stepGetClientKey, stepGetLoginServerURL,
		stepPrepareSidecarImage, stepCreatePgContainer, stepStartSidecar, stepAddMakoshRecord,
	} {
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

// TestConnectServiceToVpnHandler_FailurePath_CreateContainerFails is the
// required failure-path test: if ContainerCreate fails, the task must end up
// FAILED and no container id should ever be recorded (so rollback is a no-op).
func TestConnectServiceToVpnHandler_FailurePath_CreateContainerFails(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()

	payload := &velez_api.ConnectServiceToVpnTaskPayload{ServiceName: testServiceName}
	task := connectServiceToVpnTask(t, tasksStorage, testServiceName, payload)

	docker := newFakeDocker()
	docker.containerCreateErr = errors.New("no space left on device")
	nodeClients := newFakeNodeClients(docker)

	vpn := newFakeVpnClient()
	vpn.issuedKey = testIssuedKey
	sd := newFakeServiceDiscovery()

	registry := NewRegistry()
	registry.Register(NewConnectServiceToVpnHandler(nodeClients, vpn, sd))

	w, ok := NewTaskWorker(tasksStorage, jobsStorage, registry, "test-worker", time.Hour).(*taskWorker)
	if !ok {
		t.Fatal("expected NewTaskWorker to return *taskWorker")
	}

	err := w.run(context.Background(), task)
	if err == nil {
		t.Fatal("expected an error when the sidecar container fails to create")
	}

	finished := tasksStorage.get(task.ID)
	if finished.Status != tasks_queries.VelezTaskStatusFAILED {
		t.Errorf("expected task status FAILED, got %v", finished.Status)
	}

	if len(docker.removeCalledWith) != 0 {
		t.Errorf("expected Rollback to be a no-op (no container id was ever set), got Remove called with %v",
			docker.removeCalledWith)
	}

	if len(sd.upsertCalledWith) != 0 {
		t.Errorf("expected add_makosh_record to never run, got %d upsert calls", len(sd.upsertCalledWith))
	}
}

func proto(s string) *string {
	return &s
}
