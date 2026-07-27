package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
	"github.com/docker/go-connections/nat"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sqlc-dev/pqtype"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/ports"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testUpgradeImage    = "myimg:new"
	testNetworkAlias    = "myalias"
	testUpgradeContName = "/mysvc"
	testUpgradePgPort   = "8080/tcp"
	testCreatedID       = "created123"
	testContFixtureID   = "cont-fixture"
	testEnvKeyFoo       = "FOO"
	testEnvFoo          = "bar"
	testUpgradeSvcName  = "mysvc"
	testOldContainerID  = "old123"
	testNetworkName     = "vervnet"
)

var (
	errNetworkCreate = errors.New("network create failed")
	errNoSuchImage   = errors.New("no such image")
)

func TestUpgradeSmerdHandler_Action(t *testing.T) {
	h := NewUpgradeSmerdHandler(nil, nil, nil)

	if h.Action() != UpgradeSmerdAction {
		t.Errorf("expected action %q, got %q", UpgradeSmerdAction, h.Action())
	}
}

func TestUpgradeSmerdHandler_NewContext(t *testing.T) {
	h := NewUpgradeSmerdHandler(nil, nil, nil)

	if _, ok := h.NewContext().(*velez_api.UpgradeSmerdTaskPayload); !ok {
		t.Fatal("expected NewContext to return *velez_api.UpgradeSmerdTaskPayload")
	}
}

func TestUpgradeSmerdHandler_BuildJobs_NamesAndOrder(t *testing.T) {
	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: testUpgradeSvcName, Image: testUpgradeImage},
	}

	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)

	h := NewUpgradeSmerdHandler(nodeClients, newFakeContainerService(), newFakeConfigurationService())

	namedJobs := h.BuildJobs(payload)

	wantNames := []string{
		stepCheckSelfUpgrade, stepCaptureOldContainer, stepPrepareCreateImage, stepPauseOldContainer,
		stepCreateConfigFetcherContainer, stepGetConfigFromContainer, stepDropConfigFetcherContainer,
		stepFetchConfig, stepPrepareVervConfig, stepCreateFinalContainer, stepStartFinalContainer,
		stepHealthcheck, stepRenameOldContainer, stepDropOldContainer, stepRenameNewContainer,
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

// checkSelfUpgradeJob

func TestCheckSelfUpgradeJob_NotInsideContainer_NoOp(t *testing.T) {
	// env.GetContainerId() returns nil unless running inside an actual
	// container (cgroup/hostname probe) - true for `go test`, so this only
	// exercises the "no self-upgrade check possible" branch. The
	// self-upgrade-forbidden branch isn't reachable in this environment,
	// same limitation upgrade_steps.CheckUpgradeIsAvailable already has.
	containerService := newFakeContainerService()
	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: testUpgradeSvcName},
	}

	j := &checkSelfUpgradeJob{containerService: containerService, upgradeReq: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(containerService.inspectCalledWith) != 0 {
		t.Errorf("expected InspectSmerd not to be called, got %v", containerService.inspectCalledWith)
	}
}

// captureOldContainerJob

func TestCaptureOldContainerJob_InspectError(t *testing.T) {
	containerService := newFakeContainerService()

	containerService.inspectErr = user_errors.ErrNoSuchContainer

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: testUpgradeSvcName, Image: testUpgradeImage},
	}

	j := &captureOldContainerJob{containerService: containerService, upgradeReq: payload, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when InspectSmerd fails")
	}
}

func TestCaptureOldContainerJob_Success(t *testing.T) {
	containerService := newFakeContainerService()

	containerService.inspectResp = &velez_api.Smerd{
		Uuid:      testOldContainerID,
		Name:      testUpgradeSvcName,
		ImageName: "myimg:old",
		Ports:     []*velez_api.Port{{ServicePortNumber: 8080}},
		Volumes:   []*velez_api.Volume{{VolumeName: "data"}},
		Networks: []*velez_api.NetworkBind{
			{NetworkName: testNetworkName, Aliases: []string{testNetworkAlias, testOldContainerID}},
		},
		Env:    map[string]string{testEnvKeyFoo: testEnvFoo},
		Labels: map[string]string{"team": "core"},
	}

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: testUpgradeSvcName, Image: testUpgradeImage},
	}

	j := &captureOldContainerJob{containerService: containerService, upgradeReq: payload, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetOldContainerId() != testOldContainerID {
		t.Errorf("expected old container id 'old123', got %q", payload.GetOldContainerId())
	}

	req := payload.GetRequest()
	if req.GetName() != testUpgradeSvcName {
		t.Errorf("expected request name 'mysvc', got %q", req.GetName())
	}

	if req.GetImageName() != testUpgradeImage {
		t.Errorf("expected request image 'myimg:new' (the upgrade target, not the old image), got %q", req.GetImageName())
	}

	if req.GetEnv()[testEnvKeyFoo] != testEnvFoo {
		t.Errorf("expected env carried over from old container, got %v", req.GetEnv())
	}

	if len(req.GetSettings().GetNetwork()) != 1 || len(req.GetSettings().GetNetwork()[0].GetAliases()) != 1 {
		t.Errorf("expected the container's own uuid filtered out of network aliases, got %v", req.GetSettings().GetNetwork())
	}
}

// prepareUpgradeImageJob

func TestPrepareUpgradeImageJob_PullError(t *testing.T) {
	docker := newFakeDocker()

	docker.pullImageErr = errNoSuchImage

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Image: testUpgradeImage},
	}

	j := &prepareUpgradeImageJob{docker: docker, upgradeReq: payload, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when PullImage fails")
	}
}

func TestPrepareUpgradeImageJob_Success(t *testing.T) {
	docker := newFakeDocker()

	docker.pullImageResp = image.InspectResponse{
		Config: &dockerspec.DockerOCIImageConfig{
			ImageConfig: ocispec.ImageConfig{Labels: map[string]string{labels.MatreshkaConfigLabel: vervConfigLabelEnabled}},
		},
		RepoTags: []string{testUpgradeImage},
	}

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Image: testUpgradeImage},
	}

	j := &prepareUpgradeImageJob{docker: docker, upgradeReq: payload, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetImageLabels()[labels.MatreshkaConfigLabel] != vervConfigLabelEnabled {
		t.Errorf("expected image labels to be persisted, got %v", payload.GetImageLabels())
	}
}

// pauseOldContainerJob

func TestPauseOldContainerJob_EmptyContainerId_Error(t *testing.T) {
	payload := &velez_api.UpgradeSmerdTaskPayload{}

	j := &pauseOldContainerJob{ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error for an empty old container id")
	}
}

func TestPauseOldContainerJob_Success(t *testing.T) {
	containerAPI := newFakeContainerAPI()

	networkSettings := &container.NetworkSettings{
		Networks: map[string]*network.EndpointSettings{testNetworkName: {Aliases: []string{testNetworkAlias}}},
	}

	networkSettings.Ports = nat.PortMap{testUpgradePgPort: []nat.PortBinding{{HostPort: "40001"}}}

	containerAPI.inspectResp = container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			Name:  testUpgradeContName,
			State: &container.State{Status: container.StateRunning},
		},
		NetworkSettings: networkSettings,
	}

	payload := &velez_api.UpgradeSmerdTaskPayload{OldContainerId: strPtr(testOldContainerID)}

	job := &pauseOldContainerJob{dockerAPI: containerAPI, portManager: realPortManager(t), ctx: payload}

	err := job.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(containerAPI.pauseCalledWith) != 1 || containerAPI.pauseCalledWith[0] != testOldContainerID {
		t.Errorf("expected container paused, got %v", containerAPI.pauseCalledWith)
	}

	if len(containerAPI.networkDisconnectCalledWith) != 1 {
		t.Errorf("expected 1 network disconnect call, got %v", containerAPI.networkDisconnectCalledWith)
	}

	if len(job.portsOnHold) != 1 || job.portsOnHold[0] != 40001 {
		t.Errorf("expected port 40001 held, got %v", job.portsOnHold)
	}

	// Rollback: unpause, then reconnect any network connectToNetwork finds
	// missing. containerAPI.inspectResp is a single static fixture (no
	// per-call state), so ContainerInspect still reports testNetworkName present
	// during Rollback exactly as it did during Do() - connectToNetwork's
	// already-connected check (mirroring dockerutils.ConnectToNetwork's own
	// optimization) therefore sees it as already connected and skips the
	// NetworkConnect call. That's the fake's limitation, not the job's: this
	// only asserts what the fixture can actually support (unpause + no
	// error), not a full reconnect call.
	err = job.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected rollback error: %v", err)
	}

	if len(containerAPI.unpauseCalledWith) != 1 {
		t.Errorf("expected container unpaused on rollback, got %v", containerAPI.unpauseCalledWith)
	}
}

// renamingCreateContainerJob

func TestRenamingCreateContainerJob_Success(t *testing.T) {
	docker := newFakeDocker()

	docker.containerCreateResp = container.CreateResponse{ID: testCreatedID}

	containerAPI := newFakeContainerAPI()

	containerAPI.inspectResp = container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{ID: testCreatedID},
	}
	docker.withClient(containerAPI)

	nodeClients := newFakeNodeClients(docker)

	payload := &velez_api.UpgradeSmerdTaskPayload{
		Request: &velez_api.CreateSmerd_Request{Name: testUpgradeSvcName, Settings: &velez_api.Container_Settings{}},
	}

	j := &renamingCreateContainerJob{
		nodeClients: nodeClients,
		req:         payload,
		ctx:         payload,
		newName:     func(current string) string { return current + configFetcherContainerSuffix },
	}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetRequest().GetName() != testUpgradeSvcName+configFetcherContainerSuffix {
		t.Errorf("expected request name renamed, got %q", payload.GetRequest().GetName())
	}

	if payload.GetContainerId() != testCreatedID {
		t.Errorf("expected container id 'created123', got %q", payload.GetContainerId())
	}

	docker.removeErr = errdefs.NotFound(user_errors.ErrNetworkNotFound)

	err = j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("expected NotFound rollback error to be swallowed, got: %v", err)
	}
}

// getConfigFromScratchContainerJob

func TestGetConfigFromScratchContainerJob_EmptyContainerId_Error(t *testing.T) {
	payload := &velez_api.UpgradeSmerdTaskPayload{}

	j := &getConfigFromScratchContainerJob{imageMeta: payload, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error for an empty container id")
	}
}

func TestGetConfigFromScratchContainerJob_NonVervImage_NoOp(t *testing.T) {
	// classifyImage returns an empty systemPath for non-verv/non-postgres
	// images, so this never touches dockerAPI (left nil here on purpose).
	payload := &velez_api.UpgradeSmerdTaskPayload{
		ContainerId: strPtr("cfg123"),
		ImageLabels: map[string]string{},
	}

	j := &getConfigFromScratchContainerJob{imageMeta: payload, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetConfigFromScratchContainerJob_VervImage_ReadsConfig(t *testing.T) {
	containerAPI := newFakeContainerAPI()

	containerAPI.copyFromResp = []byte("KEY=value")

	payload := &velez_api.UpgradeSmerdTaskPayload{
		ContainerId: strPtr("cfg123"),
		ImageLabels: map[string]string{labels.MatreshkaConfigLabel: vervConfigLabelEnabled},
	}

	j := &getConfigFromScratchContainerJob{dockerAPI: containerAPI, imageMeta: payload, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// fetchUpgradeConfigJob

func TestFetchUpgradeConfigJob_RestoresNameAndMergesEnv(t *testing.T) {
	configService := newFakeConfigurationService()

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: testUpgradeSvcName},
		Request: &velez_api.CreateSmerd_Request{
			Name: testUpgradeSvcName + configFetcherContainerSuffix,
			Env:  map[string]string{},
		},
		ImageLabels: map[string]string{labels.MatreshkaConfigLabel: vervConfigLabelEnabled},
	}

	j := &fetchUpgradeConfigJob{configService: configService, upgradeReq: payload, imageMeta: payload, req: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetRequest().GetName() != testUpgradeSvcName {
		t.Errorf("expected request name restored to 'mysvc', got %q", payload.GetRequest().GetName())
	}

	if len(configService.envCalledWith) != 1 || configService.envCalledWith[0].Name != "verv_mysvc" {
		t.Errorf("expected GetEnvFromApi called with config name 'verv_mysvc', got %v", configService.envCalledWith)
	}
}

// prepareUpgradeVervConfigJob

func TestPrepareUpgradeVervConfigJob_Success(t *testing.T) {
	containerAPI := newFakeContainerAPI()

	containerAPI.networkListResp = nil // network not found -> gets created

	payload := &velez_api.UpgradeSmerdTaskPayload{
		Request: &velez_api.CreateSmerd_Request{
			Name:   testUpgradeSvcName,
			Env:    map[string]string{},
			Labels: map[string]string{},
			Settings: &velez_api.Container_Settings{
				Ports:   []*velez_api.Port{{ServicePortNumber: 8080}},
				Network: []*velez_api.NetworkBind{{NetworkName: testNetworkName}},
			},
		},
		ImageLabels: map[string]string{"custom": "label"},
	}

	pm := realPortManager(t)

	j := &prepareUpgradeVervConfigJob{dockerAPI: containerAPI, portManager: pm, imageMeta: payload, req: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetRequest().GetSettings().GetPorts()[0].GetExposedTo() == 0 {
		t.Error("expected a host port to be locked")
	}

	if payload.GetRequest().GetLabels()["custom"] != "label" {
		t.Errorf("expected image labels merged into request labels, got %v", payload.GetRequest().GetLabels())
	}

	if payload.GetRequest().GetLabels()[labels.ComposeGroupLabel] != testUpgradeSvcName {
		t.Errorf("expected compose group label set, got %v", payload.GetRequest().GetLabels())
	}

	if len(containerAPI.networkCreateCalledWith) != 1 {
		t.Errorf("expected network to be created, got %v", containerAPI.networkCreateCalledWith)
	}

	err = j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected rollback error: %v", err)
	}
}

// renameContainerJob

func TestRenameContainerJob_EmptyContainerId_Error(t *testing.T) {
	payload := &velez_api.UpgradeSmerdTaskPayload{}

	j := &renameContainerJob{ctx: payload, newName: testUpgradeSvcName}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error for an empty container id")
	}
}

func TestRenameContainerJob_SuccessAndRollback(t *testing.T) {
	containerAPI := newFakeContainerAPI()

	containerAPI.inspectResp = container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{Name: "/mysvc_new"},
	}

	payload := &velez_api.UpgradeSmerdTaskPayload{ContainerId: strPtr("new123")}

	j := &renameContainerJob{dockerAPI: containerAPI, ctx: payload, newName: testUpgradeSvcName}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(containerAPI.renameCalledWith) != 1 || containerAPI.renameCalledWith[0].newName != testUpgradeSvcName {
		t.Errorf("expected rename to 'mysvc', got %v", containerAPI.renameCalledWith)
	}

	err = j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected rollback error: %v", err)
	}

	if containerAPI.renameCalledWith[1].newName != "/mysvc_new" {
		t.Errorf("expected rollback to rename back to '/mysvc_new', got %v", containerAPI.renameCalledWith)
	}
}

// End-to-end coverage through the real handler + taskWorker.

func upgradeSmerdTask(
	t *testing.T, tasksStorage *fakeTasksStorage, entityID string, payload *velez_api.UpgradeSmerdTaskPayload,
) tasks_queries.VelezTask {
	t.Helper()

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("unexpected error marshaling payload: %v", err)
	}

	task, err := tasksStorage.CreateTask(context.Background(), tasks_queries.CreateTaskParams{
		EntityID: entityID,
		Action:   UpgradeSmerdAction,
		Context:  pqtype.NullRawMessage{RawMessage: payloadJSON, Valid: true},
	})
	if err != nil {
		t.Fatalf("unexpected error creating task: %v", err)
	}

	return task
}

// TestUpgradeSmerdHandler_HappyPath_EndToEnd drives all 15 jobs through the
// real handler + taskWorker.runJobs. docker.withClient(containerAPI) makes
// every job that calls nodeClients.Docker().Client() internally (instead of
// only jobs whose dockerAPI field was patched after BuildJobs, as in
// connect_service_to_vpn_test.go/enable_statefull_test.go) resolve to the
// same fake, since fakeDocker.Client() now returns it directly - so this can
// run via BuildJobs without any post-hoc field patching.
func TestUpgradeSmerdHandler_HappyPath_EndToEnd(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: testUpgradeSvcName, Image: testUpgradeImage},
	}
	task := upgradeSmerdTask(t, tasksStorage, testUpgradeSvcName, payload)

	containerAPI := newFakeContainerAPI()

	networkSettings := &container.NetworkSettings{
		Networks: map[string]*network.EndpointSettings{testNetworkName: {Aliases: []string{testNetworkAlias}}},
	}

	networkSettings.Ports = nat.PortMap{testUpgradePgPort: []nat.PortBinding{{HostPort: "40001"}}}

	containerAPI.inspectResp = container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:    testContFixtureID,
			Name:  testUpgradeContName,
			State: &container.State{Status: container.StateRunning},
		},
		NetworkSettings: networkSettings,
	}
	containerAPI.copyFromResp = []byte("KEY=value")

	docker := newFakeDocker()

	docker.pullImageResp = image.InspectResponse{
		Config: &dockerspec.DockerOCIImageConfig{
			ImageConfig: ocispec.ImageConfig{Labels: map[string]string{labels.MatreshkaConfigLabel: vervConfigLabelEnabled}},
		},
		RepoTags: []string{testUpgradeImage},
	}
	docker.containerCreateResp = container.CreateResponse{ID: testContFixtureID}
	docker.withClient(containerAPI)

	nodeClients := newFakeNodeClients(docker).withPortManager(realPortManager(t))

	containerService := newFakeContainerService()

	containerService.inspectResp = &velez_api.Smerd{
		Uuid:      testOldContainerID,
		Name:      testUpgradeSvcName,
		ImageName: "myimg:old",
		Ports:     []*velez_api.Port{{ServicePortNumber: 8080}},
		Networks:  []*velez_api.NetworkBind{{NetworkName: testNetworkName, Aliases: []string{testNetworkAlias}}},
		Env:       map[string]string{testEnvKeyFoo: testEnvFoo},
		Labels:    map[string]string{"team": "core"},
	}

	configService := newFakeConfigurationService()

	handler := NewUpgradeSmerdHandler(nodeClients, containerService, configService)

	taskCtx := handler.NewContext()

	err := json.Unmarshal(task.Context.RawMessage, taskCtx)
	if err != nil {
		t.Fatalf("unexpected error unmarshaling task context: %v", err)
	}

	namedJobs := handler.BuildJobs(taskCtx)

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

	finishedPayload, ok := taskCtx.(*velez_api.UpgradeSmerdTaskPayload)
	if !ok {
		t.Fatalf("expected taskCtx to be *velez_api.UpgradeSmerdTaskPayload, got %T", taskCtx)
	}

	if finishedPayload.GetOldContainerId() != testOldContainerID {
		t.Errorf("expected old container id 'old123', got %q", finishedPayload.GetOldContainerId())
	}

	// Nothing resets Request.Name after the final "_new" rename stage - only
	// the real Docker container gets renamed back to testUpgradeSvcName (rename_new_container),
	// matching do_smerd_upgrade.go's own newLaunch, which is never touched
	// again after its last SingleFunc step either.
	if finishedPayload.GetRequest().GetName() != testUpgradeSvcName+newContainerSuffix {
		t.Errorf("expected final request name 'mysvc%s', got %q", newContainerSuffix, finishedPayload.GetRequest().GetName())
	}

	for _, name := range []string{
		stepCheckSelfUpgrade, stepCaptureOldContainer, stepPrepareCreateImage, stepPauseOldContainer,
		stepCreateConfigFetcherContainer, stepGetConfigFromContainer, stepDropConfigFetcherContainer,
		stepFetchConfig, stepPrepareVervConfig, stepCreateFinalContainer, stepStartFinalContainer,
		stepHealthcheck, stepRenameOldContainer, stepDropOldContainer, stepRenameNewContainer,
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

// TestUpgradeSmerdHandler_FailurePath_NetworkCreateFails is the required
// failure-path test: prepare_verv_config's network create fails, which must
// roll back pause_old_container (unpause/reconnect) and
// create_config_fetcher_container's already-dropped container (a tolerated
// no-op remove) through the real checkpoint chain.
func TestUpgradeSmerdHandler_FailurePath_NetworkCreateFails(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: testUpgradeSvcName, Image: testUpgradeImage},
	}
	task := upgradeSmerdTask(t, tasksStorage, testUpgradeSvcName, payload)

	containerAPI := newFakeContainerAPI()

	networkSettings := &container.NetworkSettings{
		Networks: map[string]*network.EndpointSettings{testNetworkName: {Aliases: []string{testNetworkAlias}}},
	}

	networkSettings.Ports = nat.PortMap{testUpgradePgPort: []nat.PortBinding{{HostPort: "40001"}}}

	containerAPI.inspectResp = container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:    testContFixtureID,
			Name:  testUpgradeContName,
			State: &container.State{Status: container.StateRunning},
		},
		NetworkSettings: networkSettings,
	}
	containerAPI.copyFromResp = []byte("KEY=value")
	containerAPI.networkCreateErr = errNetworkCreate

	docker := newFakeDocker()

	docker.pullImageResp = image.InspectResponse{
		Config: &dockerspec.DockerOCIImageConfig{
			ImageConfig: ocispec.ImageConfig{Labels: map[string]string{labels.MatreshkaConfigLabel: vervConfigLabelEnabled}},
		},
		RepoTags: []string{testUpgradeImage},
	}
	docker.containerCreateResp = container.CreateResponse{ID: testContFixtureID}
	docker.withClient(containerAPI)

	nodeClients := newFakeNodeClients(docker).withPortManager(realPortManager(t))

	containerService := newFakeContainerService()

	containerService.inspectResp = &velez_api.Smerd{
		Uuid:     testOldContainerID,
		Name:     testUpgradeSvcName,
		Ports:    []*velez_api.Port{{ServicePortNumber: 8080}},
		Networks: []*velez_api.NetworkBind{{NetworkName: testNetworkName, Aliases: []string{testNetworkAlias}}},
		Env:      map[string]string{},
		Labels:   map[string]string{},
	}

	handler := NewUpgradeSmerdHandler(nodeClients, containerService, newFakeConfigurationService())

	taskCtx := handler.NewContext()

	err := json.Unmarshal(task.Context.RawMessage, taskCtx)
	if err != nil {
		t.Fatalf("unexpected error unmarshaling task context: %v", err)
	}

	namedJobs := handler.BuildJobs(taskCtx)

	registry := NewRegistry()
	registry.Register(handler)

	w, ok := NewTaskWorker(tasksStorage, jobsStorage, registry, "test-worker", time.Hour).(*taskWorker)
	if !ok {
		t.Fatal("expected NewTaskWorker to return *taskWorker")
	}

	runErr := w.runJobs(context.Background(), task.ID, taskCtx, namedJobs)
	if runErr == nil {
		t.Fatal("expected an error: network create fails in prepare_verv_config")
	}

	err = tasksStorage.FinishTask(context.Background(), tasks_queries.FinishTaskParams{
		ID:     task.ID,
		Status: tasks_queries.VelezTaskStatusFAILED,
		Error:  sql.NullString{String: runErr.Error(), Valid: true},
	})
	if err != nil {
		t.Fatalf("unexpected error finishing task: %v", err)
	}

	// containerAPI.inspectResp is a single static fixture, so
	// connectToNetwork's already-connected check sees testNetworkName as still
	// present during rollback and skips NetworkConnect - see
	// TestPauseOldContainerJob_Success's comment for why. unpause is the
	// meaningful, fixture-supportable assertion here.
	if len(containerAPI.unpauseCalledWith) != 1 {
		t.Errorf("expected pause_old_container to be rolled back (unpause), got %v", containerAPI.unpauseCalledWith)
	}

	finished := tasksStorage.get(task.ID)
	if finished.Status != tasks_queries.VelezTaskStatusFAILED {
		t.Errorf("expected task status FAILED, got %v", finished.Status)
	}
}

func strPtr(s string) *string { return &s }

// realPortManager returns a real, in-memory ports.PortManager (no external
// dependency beyond a local net.Listen availability probe on the given
// port) - no hand-written PortManager fake exists in this repo, and this one
// is cheap/pure enough to use directly in unit tests.
func realPortManager(t *testing.T) node_clients.PortManager {
	t.Helper()

	return ports.NewPortManager([]int{58080}, nil)
}
