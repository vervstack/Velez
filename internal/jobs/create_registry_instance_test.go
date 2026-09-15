package jobs

import (
	"encoding/json"
	"testing"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	testRegistryInstanceName = "my-registry"
)

func TestCreateRegistryInstanceHandler_Action(t *testing.T) {
	h := NewCreateRegistryInstanceHandler(nil, nil, nil, nil, nil)

	if h.Action() != CreateRegistryInstanceAction {
		t.Errorf("expected action %q, got %q", CreateRegistryInstanceAction, h.Action())
	}
}

func TestCreateRegistryInstanceHandler_NewContext(t *testing.T) {
	h := NewCreateRegistryInstanceHandler(nil, nil, nil, nil, nil)

	if _, ok := h.NewContext().(*velez_api.CreateRegistryInstanceTaskPayload); !ok {
		t.Fatal("expected NewContext to return *velez_api.CreateRegistryInstanceTaskPayload")
	}
}

func TestCreateRegistryInstanceHandler_BuildJobs_NamesAndOrder(t *testing.T) {
	payload := &velez_api.CreateRegistryInstanceTaskPayload{
		Request: &velez_api.CreateRegistryInstance_Request{Name: testRegistryInstanceName},
	}

	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)

	clusterStorage := &fakeClusterStorage{}
	storageContainer := storage.NewStorageContainer(clusterStorage)

	h := NewCreateRegistryInstanceHandler(nodeClients, newFakeRuntimes(docker, nil), storageContainer, nil, nil)

	namedJobs := h.BuildJobs(payload)

	wantNames := []string{
		stepGenerateCredentials, stepPutSecret, stepResolvePorts, stepCreateLoaderContainer, stepStartSidecar,
		stepWriteHtpasswd, stepDropContainer, stepDeployRegistry, stepDeployRegistryUi,
		stepRegisterRegistryInstance, stepRegisterRegistryRow,
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

// bind_owner_resource is only appended when the request names an owner
// service - see BuildJobs's doc comment.
func TestCreateRegistryInstanceHandler_BuildJobs_OwnerService_AppendsBindStep(t *testing.T) {
	ownerService := "my-app"
	payload := &velez_api.CreateRegistryInstanceTaskPayload{
		Request: &velez_api.CreateRegistryInstance_Request{Name: testRegistryInstanceName, OwnerService: &ownerService},
	}

	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)

	clusterStorage := &fakeClusterStorage{}
	storageContainer := storage.NewStorageContainer(clusterStorage)

	h := NewCreateRegistryInstanceHandler(nodeClients, newFakeRuntimes(docker, nil), storageContainer, nil, nil)

	namedJobs := h.BuildJobs(payload)

	last := namedJobs[len(namedJobs)-1]
	if last.Name != stepBindOwnerResource {
		t.Errorf("expected last job named %q, got %q", stepBindOwnerResource, last.Name)
	}
}

// CreateRegistryInstanceTaskPayload is persisted to velez.tasks as plain
// encoding/json, not protojson (internal/jobs' TaskContext contract). Unlike
// CreateSmerdTaskPayload/UpgradeSmerdTaskPayload, none of its fields - nor
// CreateRegistryInstance_Request's - are a oneof-backed Go interface field,
// so no hand-written MarshalJSON/UnmarshalJSON is needed. This is the
// throwaway proof required by CLAUDE.md's jobs-engine hard-won rule before
// wiring a job payload for real.
func TestCreateRegistryInstanceTaskPayload_JsonRoundTrip(t *testing.T) {
	environment := "prod"
	box := "box-1"
	exposeToPort := uint32(15000)
	ownerService := "my-app"
	username := "verv"
	password := "s3cr3t"
	containerID := "container-123"
	exposedPort := uint32(15001)
	uiExposedPort := uint32(15002)

	original := &velez_api.CreateRegistryInstanceTaskPayload{
		Request: &velez_api.CreateRegistryInstance_Request{
			Name:         testRegistryInstanceName,
			Environment:  &environment,
			Box:          &box,
			ExposeToPort: &exposeToPort,
			OwnerService: &ownerService,
		},
		Username:      &username,
		Password:      &password,
		ContainerId:   &containerID,
		ExposedPort:   &exposedPort,
		UiExposedPort: &uiExposedPort,
	}

	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("unexpected error marshaling payload: %v", err)
	}

	var roundTripped velez_api.CreateRegistryInstanceTaskPayload

	err = json.Unmarshal(raw, &roundTripped)
	if err != nil {
		t.Fatalf("unexpected error unmarshaling payload: %v", err)
	}

	if roundTripped.GetRequest().GetName() != testRegistryInstanceName {
		t.Errorf("expected request.name %q, got %q", testRegistryInstanceName, roundTripped.GetRequest().GetName())
	}

	if roundTripped.GetRequest().GetEnvironment() != environment {
		t.Errorf("expected request.environment %q, got %q", environment, roundTripped.GetRequest().GetEnvironment())
	}

	if roundTripped.GetRequest().GetBox() != box {
		t.Errorf("expected request.box %q, got %q", box, roundTripped.GetRequest().GetBox())
	}

	if roundTripped.GetRequest().GetExposeToPort() != exposeToPort {
		t.Errorf(
			"expected request.expose_to_port %d, got %d", exposeToPort, roundTripped.GetRequest().GetExposeToPort(),
		)
	}

	if roundTripped.GetRequest().GetOwnerService() != ownerService {
		t.Errorf(
			"expected request.owner_service %q, got %q", ownerService, roundTripped.GetRequest().GetOwnerService(),
		)
	}

	if roundTripped.GetUsername() != username {
		t.Errorf("expected username %q, got %q", username, roundTripped.GetUsername())
	}

	if roundTripped.GetPassword() != password {
		t.Errorf("expected password %q, got %q", password, roundTripped.GetPassword())
	}

	if roundTripped.GetContainerId() != containerID {
		t.Errorf("expected container_id %q, got %q", containerID, roundTripped.GetContainerId())
	}

	if roundTripped.GetExposedPort() != exposedPort {
		t.Errorf("expected exposed_port %d, got %d", exposedPort, roundTripped.GetExposedPort())
	}

	if roundTripped.GetUiExposedPort() != uiExposedPort {
		t.Errorf("expected ui_exposed_port %d, got %d", uiExposedPort, roundTripped.GetUiExposedPort())
	}
}

func TestRegistryaasDataVolumeName(t *testing.T) {
	got := registryaasDataVolumeName(testRegistryInstanceName)
	want := "my-registry-data"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestRegistryaasAuthVolumeName(t *testing.T) {
	got := registryaasAuthVolumeName(testRegistryInstanceName)
	want := "my-registry-auth"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestRegistryaasUiServiceName(t *testing.T) {
	got := registryaasUiServiceName(testRegistryInstanceName)
	want := "my-registry-ui"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestRegistryInstanceSecretRef(t *testing.T) {
	got := registryInstanceSecretRef(testRegistryInstanceName)
	want := domain.SecretRef{Scope: "registryaas", Owner: testRegistryInstanceName, Key: "password"}

	if got != want {
		t.Errorf("expected %+v, got %+v", want, got)
	}
}

func TestRegistryInternalUrl(t *testing.T) {
	got := registryInternalUrl(testRegistryInstanceName)
	want := "http://my-registry:5000"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

// registryInstanceUrl's localhost branch only runs when velez itself is not
// in a container - mirrors enable_statefull_test.go's getRootDsnJob tests'
// use of env.IsInContainer() to skip the branch that can't apply here.
func TestRegistryInstanceUrl_NotInContainer_UsesLocalhostAndExposedPort(t *testing.T) {
	if env.IsInContainer() {
		t.Skip("the localhost+exposed-port branch only runs when velez is not itself in a container")
	}

	got := registryInstanceUrl(testRegistryInstanceName, 15001)
	want := "http://localhost:15001"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestHtpasswdLine(t *testing.T) {
	line, err := htpasswdLine("verv", "s3cr3t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(line) == 0 || line[len(line)-1] != '\n' {
		t.Fatalf("expected line to end with a newline, got %q", line)
	}

	prefix := "verv:$2a$"
	if len(line) < len(prefix) || line[:len(prefix)] != prefix {
		t.Errorf("expected line to start with %q (bcrypt hash), got %q", prefix, line)
	}
}
