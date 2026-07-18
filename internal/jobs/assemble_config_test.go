package jobs

import (
	"context"
	"errors"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/errdefs"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/tests/config_mocks"
)

func TestAssembleConfigHandler_Action(t *testing.T) {
	h := NewAssembleConfigHandler(nil)

	if h.Action() != AssembleConfigAction {
		t.Errorf("expected action %q, got %q", AssembleConfigAction, h.Action())
	}
}

func TestAssembleConfigHandler_NewContext(t *testing.T) {
	h := NewAssembleConfigHandler(nil)

	if _, ok := h.NewContext().(*velez_api.AssembleConfigTaskPayload); !ok {
		t.Fatal("expected NewContext to return *velez_api.AssembleConfigTaskPayload")
	}
}

func TestClassifyImage(t *testing.T) {
	cases := []struct {
		name           string
		imageLabels    map[string]string
		imageTags      []string
		wantConfType   string
		wantFormat     velez_api.ConfigFormat
		wantSystemPath string
	}{
		{
			name:           "matreshka labeled image",
			imageLabels:    map[string]string{labels.MatreshkaConfigLabel: "true"},
			wantConfType:   confTypeVerv,
			wantFormat:     velez_api.ConfigFormat_env,
			wantSystemPath: "/app/config/config.yaml",
		},
		{
			name:         "postgres by tag",
			imageTags:    []string{"docker.io/library/postgres:16"},
			wantConfType: confTypePg,
			wantFormat:   velez_api.ConfigFormat_env,
		},
		{
			name:         "plain image",
			imageTags:    []string{"docker.io/library/nginx:1"},
			wantConfType: confTypePlain,
			wantFormat:   velez_api.ConfigFormat_env,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			confType, format, systemPath := classifyImage(tc.imageLabels, tc.imageTags)

			if confType != tc.wantConfType {
				t.Errorf("expected confType %q, got %q", tc.wantConfType, confType)
			}
			if format != tc.wantFormat {
				t.Errorf("expected format %v, got %v", tc.wantFormat, format)
			}
			if systemPath != tc.wantSystemPath {
				t.Errorf("expected systemPath %q, got %q", tc.wantSystemPath, systemPath)
			}
		})
	}
}

func TestParseConfigJob_NoContentRaw(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{}

	j := &parseConfigJob{req: payload, content: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(payload.GetContent()) != 0 {
		t.Errorf("expected no content, got %q", payload.GetContent())
	}
}

func TestParseConfigJob_YamlPlain(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{}
	payload.SetConfType(confTypePlain)
	payload.SetConfigFormat(velez_api.ConfigFormat_yaml)
	payload.SetContentRaw([]byte("foo: bar\n"))

	j := &parseConfigJob{req: payload, content: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(payload.GetContent()) != "FOO=bar\n" {
		t.Errorf("expected content %q, got %q", "FOO=bar\n", payload.GetContent())
	}
}

func TestParseConfigJob_YamlMatreshka(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{}
	payload.SetConfType(confTypeVerv)
	payload.SetConfigFormat(velez_api.ConfigFormat_yaml)
	payload.SetContentRaw(config_mocks.HelloWorld)

	j := &parseConfigJob{req: payload, content: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(payload.GetContent()) == 0 {
		t.Error("expected non-empty content")
	}
}

// EnvFormat pins the current (pre-existing, inherited from
// do_assemble_config.go's getResult) behavior of evon.Unmarshal(raw, &node):
// passed a nil *evon.Node destination it returns no error but never
// populates the node, so ConfigFormat_env content is silently dropped here
// exactly as it is in the pipeline version. Not something this migration
// fixes - see PR discussion before changing it.
func TestParseConfigJob_EnvFormat(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{}
	payload.SetConfType(confTypePlain)
	payload.SetConfigFormat(velez_api.ConfigFormat_env)
	payload.SetContentRaw([]byte("FOO=bar\n"))

	j := &parseConfigJob{req: payload, content: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(payload.GetContent()) != 0 {
		t.Errorf("expected no content (pre-existing evon.Unmarshal no-op), got %q", payload.GetContent())
	}
}

// prepareScratchImageJob depends only on the narrow node_clients.Docker
// interface, so it's fully unit-testable with fakeDocker (fakes_test.go).

func TestPrepareScratchImageJob_Success(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{
		ImageName: "nginx:1.25",
	}

	imageConfig := &dockerspec.DockerOCIImageConfig{
		ImageConfig: ocispec.ImageConfig{
			Labels: map[string]string{"foo": "bar"},
		},
	}

	docker := newFakeDocker()
	docker.pullImageResp = image.InspectResponse{
		RepoTags: []string{"nginx:1.25"},
		Config:   imageConfig,
	}

	j := &prepareScratchImageJob{docker: docker, req: payload, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetImageLabels()["foo"] != "bar" {
		t.Errorf("expected image label foo=bar, got %v", payload.GetImageLabels())
	}
	if len(payload.GetImageTags()) != 1 || payload.GetImageTags()[0] != "nginx:1.25" {
		t.Errorf("expected image tags [nginx:1.25], got %v", payload.GetImageTags())
	}
}

// TestPrepareScratchImageJob_PullImageError is the migration checklist's
// required failure-path test for this job: PullImage's error must propagate
// out of Do, and no image metadata should be recorded on the task context.
func TestPrepareScratchImageJob_PullImageError(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{
		ImageName: "nginx:1.25",
	}

	docker := newFakeDocker()
	docker.pullImageErr = errors.New("registry unreachable")

	j := &prepareScratchImageJob{docker: docker, req: payload, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when PullImage fails")
	}

	if payload.GetImageLabels() != nil {
		t.Errorf("expected no image labels set on failure, got %v", payload.GetImageLabels())
	}
}

// createScratchContainerJob depends on the full node_clients.NodeClients
// container (only via its Docker() method), covered here with
// fakeNodeClients wrapping fakeDocker.

func TestCreateScratchContainerJob_Success(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{
		ServiceName: "my-svc",
		ImageName:   "nginx:1.25",
	}

	docker := newFakeDocker()
	docker.containerCreateResp = container.CreateResponse{ID: "abc123"}

	nodeClients := newFakeNodeClients(docker)

	j := &createScratchContainerJob{nodeClients: nodeClients, req: payload, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetContainerId() != "abc123" {
		t.Errorf("expected container id 'abc123', got %q", payload.GetContainerId())
	}
}

// TestCreateScratchContainerJob_ContainerCreateError is the required
// failure-path test for this job: ContainerCreate's error must propagate,
// and no container id should be recorded.
func TestCreateScratchContainerJob_ContainerCreateError(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{
		ServiceName: "my-svc",
		ImageName:   "nginx:1.25",
	}

	docker := newFakeDocker()
	docker.containerCreateErr = errors.New("no space left on device")

	nodeClients := newFakeNodeClients(docker)

	j := &createScratchContainerJob{nodeClients: nodeClients, req: payload, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when ContainerCreate fails")
	}

	if payload.GetContainerId() != "" {
		t.Errorf("expected no container id set on failure, got %q", payload.GetContainerId())
	}
}

func TestCreateScratchContainerJob_Rollback_NoContainerId_NoOp(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{}

	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)

	j := &createScratchContainerJob{nodeClients: nodeClients, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docker.removeCalledWith) != 0 {
		t.Errorf("expected Remove not to be called, got %v", docker.removeCalledWith)
	}
}

func TestCreateScratchContainerJob_Rollback_RemovesContainer(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{}
	payload.SetContainerId("abc123")

	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)

	j := &createScratchContainerJob{nodeClients: nodeClients, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docker.removeCalledWith) != 1 || docker.removeCalledWith[0] != "abc123" {
		t.Errorf("expected Remove called with 'abc123', got %v", docker.removeCalledWith)
	}
}

// TestCreateScratchContainerJob_Rollback_RemoveErrorPropagates covers the
// Rollback failure path: a generic (non-NotFound) Remove error must
// propagate out of Rollback rather than being swallowed.
func TestCreateScratchContainerJob_Rollback_RemoveErrorPropagates(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{}
	payload.SetContainerId("abc123")

	docker := newFakeDocker()
	docker.removeErr = errors.New("docker daemon unreachable")

	nodeClients := newFakeNodeClients(docker)

	j := &createScratchContainerJob{nodeClients: nodeClients, ctx: payload}

	err := j.Rollback(context.Background())
	if err == nil {
		t.Fatal("expected an error when Remove fails")
	}
}

func TestCreateScratchContainerJob_Rollback_NotFoundIsSwallowed(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{}
	payload.SetContainerId("abc123")

	docker := newFakeDocker()
	docker.removeErr = errdefs.NotFound(errors.New("no such container"))

	nodeClients := newFakeNodeClients(docker)

	j := &createScratchContainerJob{nodeClients: nodeClients, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("expected NotFound error to be swallowed, got: %v", err)
	}
}

// fetchConfigJob's dockerAPI field is a client.APIClient - the full Docker
// engine SDK client interface (100+ methods across container/image/network/
// volume/system sub-clients). Hand-writing a fake for it is out of scope
// here (and out of proportion to the payoff); see the final report for the
// one branch that stays uncovered at unit level as a result. Both cases
// below reach real, meaningful behavior without ever touching dockerAPI:
// the empty-container-id guard clause, and any image classification whose
// systemPath is empty (dockerAPI is nil in both jobs below and is never
// dereferenced).

// TestFetchConfigJob_EmptyContainerId_Error is the required failure-path
// test for this job.
func TestFetchConfigJob_EmptyContainerId_Error(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{
		ServiceName: "my-svc",
	}

	j := &fetchConfigJob{
		req:       payload,
		imageMeta: payload,
		container: payload,
		ctx:       payload,
		content:   payload,
	}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error for an empty container id")
	}
}

func TestFetchConfigJob_PlainImage_SkipsDockerRead(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{
		ServiceName: "my-svc",
	}
	payload.SetContainerId("abc123")
	payload.SetImageTags([]string{"docker.io/library/nginx:1"})

	j := &fetchConfigJob{
		req:       payload,
		imageMeta: payload,
		container: payload,
		ctx:       payload,
		content:   payload,
	}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetConfType() != confTypePlain {
		t.Errorf("expected conf type %q, got %q", confTypePlain, payload.GetConfType())
	}
	if len(payload.GetContentRaw()) != 0 {
		t.Errorf("expected no content raw fetched, got %q", payload.GetContentRaw())
	}
}

// dropScratchContainerJob depends only on the narrow node_clients.Docker
// interface, so it's fully unit-testable with fakeDocker.

func TestDropScratchContainerJob_NoContainerId_NoOp(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{}

	docker := newFakeDocker()

	j := &dropScratchContainerJob{docker: docker, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docker.removeCalledWith) != 0 {
		t.Errorf("expected Remove not to be called, got %v", docker.removeCalledWith)
	}
}

func TestDropScratchContainerJob_Success(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{}
	payload.SetContainerId("abc123")

	docker := newFakeDocker()

	j := &dropScratchContainerJob{docker: docker, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docker.removeCalledWith) != 1 || docker.removeCalledWith[0] != "abc123" {
		t.Errorf("expected Remove called with 'abc123', got %v", docker.removeCalledWith)
	}
}

// TestDropScratchContainerJob_RemoveErrorPropagates is the required
// failure-path test for this job: unlike createScratchContainerJob's
// Rollback, dropScratchContainerJob.Do wraps and propagates every Remove
// error, including NotFound.
func TestDropScratchContainerJob_RemoveErrorPropagates(t *testing.T) {
	payload := &velez_api.AssembleConfigTaskPayload{}
	payload.SetContainerId("abc123")

	docker := newFakeDocker()
	docker.removeErr = errors.New("docker daemon unreachable")

	j := &dropScratchContainerJob{docker: docker, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when Remove fails")
	}
}
