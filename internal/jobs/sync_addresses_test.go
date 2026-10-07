package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	testRootService = "root_service"
	testWebUiName   = "root_service-web-ui"
	testS3Instance  = "s3main"
)

var errAddressBookDown = errors.New("address book down")

func newCreateSmerdPayload(containerLabels map[string]string) *velez_api.CreateSmerdTaskPayload {
	request := &velez_api.CreateSmerd_Request{
		Name:   testWebUiName,
		Labels: containerLabels,
	}

	payload := &velez_api.CreateSmerdTaskPayload{}
	payload.SetRequest(request)

	return payload
}

func Test_CreateSmerdHandler_BuildJobs_EndsWithSyncAddresses(t *testing.T) {
	nodeClients := newFakeNodeClients(newFakeDocker())
	h := NewCreateSmerdHandler(nodeClients, nil, nil, newFakeAddressBook())

	namedJobs := h.BuildJobs(newCreateSmerdPayload(nil))

	last := namedJobs[len(namedJobs)-1]
	if last.Name != stepSyncAddresses {
		t.Fatalf("expected last job %q, got %q", stepSyncAddresses, last.Name)
	}

	if namedJobs[len(namedJobs)-2].Name != "subscribe_for_config_changes" {
		t.Errorf("expected sync_addresses right after the config subscription, got %q",
			namedJobs[len(namedJobs)-2].Name)
	}
}

func Test_CreateSmerdSyncAddresses_SyncsOwnServiceLabel(t *testing.T) {
	addressBook := newFakeAddressBook()
	payload := newCreateSmerdPayload(map[string]string{labels.VervServiceLabel: testRootService})

	job := &syncAddressesJob{addressBook: addressBook, root: smerdRootService{req: payload}}

	err := job.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(addressBook.syncCalledWith) != 1 || addressBook.syncCalledWith[0] != testRootService {
		t.Errorf("expected sync for %q, got %v", testRootService, addressBook.syncCalledWith)
	}
}

func Test_CreateSmerdSyncAddresses_WebUiSidecarSyncsServedService(t *testing.T) {
	addressBook := newFakeAddressBook()
	payload := newCreateSmerdPayload(map[string]string{
		labels.VervServiceLabel: testWebUiName,
		labels.WebUiForLabel:    testRootService,
	})

	job := &syncAddressesJob{addressBook: addressBook, root: smerdRootService{req: payload}}

	err := job.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(addressBook.syncCalledWith) != 1 || addressBook.syncCalledWith[0] != testRootService {
		t.Errorf("expected sync for %q, got %v", testRootService, addressBook.syncCalledWith)
	}
}

func Test_CreateSmerdSyncAddresses_NoServiceLabelSkipsSync(t *testing.T) {
	addressBook := newFakeAddressBook()
	payload := newCreateSmerdPayload(map[string]string{})

	job := &syncAddressesJob{addressBook: addressBook, root: smerdRootService{req: payload}}

	err := job.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(addressBook.syncCalledWith) != 0 {
		t.Errorf("expected no sync, got %v", addressBook.syncCalledWith)
	}
}

func Test_CreateSmerdSyncAddresses_SyncErrorDoesNotFailJob(t *testing.T) {
	addressBook := newFakeAddressBook()

	addressBook.syncErr = errAddressBookDown

	payload := newCreateSmerdPayload(map[string]string{labels.VervServiceLabel: testRootService})

	job := &syncAddressesJob{addressBook: addressBook, root: smerdRootService{req: payload}}

	err := job.Do(context.Background())
	if err != nil {
		t.Fatalf("expected a sync error to be swallowed, got %v", err)
	}

	if len(addressBook.syncCalledWith) != 1 {
		t.Errorf("expected 1 sync attempt, got %v", addressBook.syncCalledWith)
	}
}

func Test_ConnectServiceToVpnSyncAddresses_SyncsConnectedService(t *testing.T) {
	addressBook := newFakeAddressBook()
	payload := &velez_api.ConnectServiceToVpnTaskPayload{ServiceName: testServiceName}
	h := NewConnectServiceToVpnHandler(
		newFakeNodeClients(newFakeDocker()), newFakeVpnClient(), newFakeServiceDiscovery(), nil, addressBook)

	namedJobs := h.BuildJobs(payload)

	last := namedJobs[len(namedJobs)-1]
	if last.Name != stepSyncAddresses {
		t.Fatalf("expected last job %q, got %q", stepSyncAddresses, last.Name)
	}

	err := last.Job.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(addressBook.syncCalledWith) != 1 || addressBook.syncCalledWith[0] != testServiceName {
		t.Errorf("expected sync for %q, got %v", testServiceName, addressBook.syncCalledWith)
	}
}

func Test_ConnectServiceToVpnSyncAddresses_SyncErrorDoesNotFailJob(t *testing.T) {
	addressBook := newFakeAddressBook()

	addressBook.syncErr = errAddressBookDown

	job := &syncAddressesJob{addressBook: addressBook, root: staticRootService(testServiceName)}

	err := job.Do(context.Background())
	if err != nil {
		t.Fatalf("expected a sync error to be swallowed, got %v", err)
	}
}

func runDropTaskWithLabels(
	t *testing.T, addressBook *fakeAddressBook, containerLabels map[string]string,
) tasks_queries.VelezTask {
	t.Helper()

	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()
	docker := newFakeDocker()

	containerAPI := newFakeContainerAPI()

	containerAPI.inspectResp = container.InspectResponse{Config: &container.Config{Labels: containerLabels}}
	docker.withClient(containerAPI)

	req := &velez_api.DropSmerd_Request{Uuids: []string{testUUID1}}
	task := dropSmerdTask(t, tasksStorage, "batch-sync", req)

	registry := NewRegistry()
	registry.Register(NewDropSmerdHandler(newFakeRuntimes(docker, nil), addressBook))

	w := NewTaskWorker(tasksStorage, jobsStorage, registry, "test-worker", time.Hour, 1)

	err := w.run(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error running task: %v", err)
	}

	return tasksStorage.get(task.ID)
}

func Test_DropSmerd_SyncsRootServiceOfDroppedContainer(t *testing.T) {
	addressBook := newFakeAddressBook()

	finished := runDropTaskWithLabels(t, addressBook, map[string]string{
		labels.VervServiceLabel: testWebUiName,
		labels.WebUiForLabel:    testRootService,
	})

	if finished.Status != tasks_queries.VelezTaskStatusDONE {
		t.Fatalf("expected task status DONE, got %v", finished.Status)
	}

	if len(addressBook.syncCalledWith) != 1 || addressBook.syncCalledWith[0] != testRootService {
		t.Errorf("expected sync for %q, got %v", testRootService, addressBook.syncCalledWith)
	}
}

func Test_DropSmerd_SyncErrorDoesNotFailTask(t *testing.T) {
	addressBook := newFakeAddressBook()

	addressBook.syncErr = errAddressBookDown

	finished := runDropTaskWithLabels(t, addressBook, map[string]string{labels.VervServiceLabel: testRootService})

	if finished.Status != tasks_queries.VelezTaskStatusDONE {
		t.Fatalf("expected task status DONE, got %v", finished.Status)
	}

	var result velez_api.DropSmerdTaskPayload

	err := json.Unmarshal(finished.Context.RawMessage, &result)
	if err != nil {
		t.Fatalf("unexpected error unmarshaling task context: %v", err)
	}

	if len(result.GetFailed()) != 0 || len(result.GetSuccessful()) != 1 {
		t.Errorf("expected the drop to still be reported successful, got %v / %v",
			result.GetSuccessful(), result.GetFailed())
	}
}

func Test_DropSmerd_UnlabelledContainerSkipsSync(t *testing.T) {
	addressBook := newFakeAddressBook()

	runDropTaskWithLabels(t, addressBook, map[string]string{})

	if len(addressBook.syncCalledWith) != 0 {
		t.Errorf("expected no sync, got %v", addressBook.syncCalledWith)
	}
}

func runLastJob(t *testing.T, namedJobs []NamedJob) {
	t.Helper()

	last := namedJobs[len(namedJobs)-1]
	if last.Name != stepSyncAddresses {
		t.Fatalf("expected last job %q, got %q", stepSyncAddresses, last.Name)
	}

	err := last.Job.Do(context.Background())
	if err != nil {
		t.Fatalf("expected sync_addresses never to fail, got %v", err)
	}
}

func assertSyncedOnce(t *testing.T, addressBook *fakeAddressBook, want string) {
	t.Helper()

	if len(addressBook.syncCalledWith) != 1 || addressBook.syncCalledWith[0] != want {
		t.Errorf("expected one sync for %q, got %v", want, addressBook.syncCalledWith)
	}
}

func newCreateS3InstancePayload(isWebUiEnabled bool) *velez_api.CreateS3InstanceTaskPayload {
	request := &velez_api.CreateS3Instance_Request{Name: testS3Instance, EnableWebUi: isWebUiEnabled}

	return &velez_api.CreateS3InstanceTaskPayload{Request: request}
}

func newCreateS3InstanceHandlerWith(addressBook *fakeAddressBook) TaskHandler {
	return NewCreateS3InstanceHandler(
		newFakeNodeClients(newFakeDocker()), &fakeClusterStorage{}, nil, nil, nil, nil, nil, addressBook)
}

func Test_CreateS3Instance_SyncsServiceAfterWebUi(t *testing.T) {
	addressBook := newFakeAddressBook()
	h := newCreateS3InstanceHandlerWith(addressBook)

	namedJobs := h.BuildJobs(newCreateS3InstancePayload(true))

	if namedJobs[len(namedJobs)-2].Name != stepWaitGarageWebUi {
		t.Errorf("expected sync_addresses right after %q, got %q",
			stepWaitGarageWebUi, namedJobs[len(namedJobs)-2].Name)
	}

	runLastJob(t, namedJobs)
	assertSyncedOnce(t, addressBook, domain.S3ServiceName(testS3Instance))
}

func Test_CreateS3Instance_SyncsServiceWithoutWebUi(t *testing.T) {
	addressBook := newFakeAddressBook()
	h := newCreateS3InstanceHandlerWith(addressBook)

	namedJobs := h.BuildJobs(newCreateS3InstancePayload(false))

	if namedJobs[len(namedJobs)-2].Name != stepApplyGarageLayout {
		t.Errorf("expected sync_addresses right after %q, got %q",
			stepApplyGarageLayout, namedJobs[len(namedJobs)-2].Name)
	}

	runLastJob(t, namedJobs)
	assertSyncedOnce(t, addressBook, domain.S3ServiceName(testS3Instance))
}

func Test_CreateS3Instance_SyncErrorDoesNotFailJob(t *testing.T) {
	addressBook := newFakeAddressBook()

	addressBook.syncErr = errAddressBookDown

	h := newCreateS3InstanceHandlerWith(addressBook)

	namedJobs := h.BuildJobs(newCreateS3InstancePayload(true))

	runLastJob(t, namedJobs)

	if len(addressBook.syncCalledWith) != 1 {
		t.Errorf("expected 1 sync attempt, got %v", addressBook.syncCalledWith)
	}
}

func newUpgradeSmerdPayload(containerLabels map[string]string) *velez_api.UpgradeSmerdTaskPayload {
	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: testUpgradeSvcName, Image: testUpgradeImage},
	}

	request := &velez_api.CreateSmerd_Request{Name: testUpgradeSvcName, Labels: containerLabels}
	payload.SetRequest(request)

	return payload
}

func Test_UpgradeSmerd_SyncsRootServiceAfterSidecars(t *testing.T) {
	addressBook := newFakeAddressBook()
	h := NewUpgradeSmerdHandler(
		newFakeNodeClients(newFakeDocker()), newFakeContainerService(), newFakeConfigurationService(), nil,
		addressBook)

	namedJobs := h.BuildJobs(newUpgradeSmerdPayload(map[string]string{labels.VervServiceLabel: testRootService}))

	if namedJobs[len(namedJobs)-2].Name != stepRecreateSidecars {
		t.Errorf("expected sync_addresses right after %q, got %q",
			stepRecreateSidecars, namedJobs[len(namedJobs)-2].Name)
	}

	runLastJob(t, namedJobs)
	assertSyncedOnce(t, addressBook, testRootService)
}

func Test_UpgradeSmerd_WebUiSidecarSyncsServedService(t *testing.T) {
	addressBook := newFakeAddressBook()
	h := NewUpgradeSmerdHandler(
		newFakeNodeClients(newFakeDocker()), newFakeContainerService(), newFakeConfigurationService(), nil,
		addressBook)

	namedJobs := h.BuildJobs(newUpgradeSmerdPayload(map[string]string{
		labels.VervServiceLabel: testWebUiName,
		labels.WebUiForLabel:    testRootService,
	}))

	runLastJob(t, namedJobs)
	assertSyncedOnce(t, addressBook, testRootService)
}

func Test_UpgradeSmerd_SyncErrorDoesNotFailJob(t *testing.T) {
	addressBook := newFakeAddressBook()

	addressBook.syncErr = errAddressBookDown

	h := NewUpgradeSmerdHandler(
		newFakeNodeClients(newFakeDocker()), newFakeContainerService(), newFakeConfigurationService(), nil,
		addressBook)

	namedJobs := h.BuildJobs(newUpgradeSmerdPayload(map[string]string{labels.VervServiceLabel: testRootService}))

	runLastJob(t, namedJobs)

	if len(addressBook.syncCalledWith) != 1 {
		t.Errorf("expected 1 sync attempt, got %v", addressBook.syncCalledWith)
	}
}

func Test_RegisterContainer_SyncsRegisteredServiceLast(t *testing.T) {
	addressBook := newFakeAddressBook()
	h := NewRegisterContainerHandler(&fakeClusterStorage{}, nil, nil, nil, addressBook)
	payload := &velez_api.RegisterContainerTaskPayload{ServiceName: testRootService}

	namedJobs := h.BuildJobs(payload)

	runLastJob(t, namedJobs)
	assertSyncedOnce(t, addressBook, testRootService)
}

func Test_RegisterContainer_SyncErrorDoesNotFailJob(t *testing.T) {
	addressBook := newFakeAddressBook()

	addressBook.syncErr = errAddressBookDown

	h := NewRegisterContainerHandler(&fakeClusterStorage{}, nil, nil, nil, addressBook)
	payload := &velez_api.RegisterContainerTaskPayload{ServiceName: testRootService}

	namedJobs := h.BuildJobs(payload)

	runLastJob(t, namedJobs)

	if len(addressBook.syncCalledWith) != 1 {
		t.Errorf("expected 1 sync attempt, got %v", addressBook.syncCalledWith)
	}
}
