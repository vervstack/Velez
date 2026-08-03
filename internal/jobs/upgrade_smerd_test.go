package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
	"github.com/docker/go-connections/nat"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sqlc-dev/pqtype"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/clients/node_clients/ports"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/environments"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
	"go.vervstack.ru/Velez/tests/test_helper"
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
	testUpgradeEnv      = "STAGE"
)

var (
	errNetworkCreate = errors.New("network create failed")
	errNoSuchImage   = errors.New("no such image")
)

func TestUpgradeSmerdHandler_Action(t *testing.T) {
	h := NewUpgradeSmerdHandler(nil, nil, nil, nil)

	if h.Action() != UpgradeSmerdAction {
		t.Errorf("expected action %q, got %q", UpgradeSmerdAction, h.Action())
	}
}

func TestUpgradeSmerdHandler_NewContext(t *testing.T) {
	h := NewUpgradeSmerdHandler(nil, nil, nil, nil)

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

	h := NewUpgradeSmerdHandler(nodeClients, newFakeContainerService(), newFakeConfigurationService(), nil)

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
	t.Parallel()

	// env.GetContainerId() returns nil unless running inside an actual
	// container (cgroup/hostname probe) - true for `go test`, so this only
	// exercises the "no self-upgrade check possible" branch: Do returns
	// before ever touching containerService/runtimes, so this mainly proves
	// the real objects wire together. The self-upgrade-forbidden branch
	// isn't reachable in this environment, same limitation
	// upgrade_steps.CheckUpgradeIsAvailable already has.
	containerService, runtimes, _ := newRealUpgradeFixture(t, testUpgradeEnv)

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: testUpgradeSvcName, Environment: testUpgradeEnv},
	}

	j := &checkSelfUpgradeJob{containerService: containerService, upgradeReq: payload, runtimes: runtimes}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// captureOldContainerJob

func TestCaptureOldContainerJob_InspectError(t *testing.T) {
	t.Parallel()

	// runtimes stays nil (unchanged shape): no container is ever created
	// under this name, so containerService.InspectSmerd gets a real 404 from
	// the daemon, and with no fallback resolver that error propagates
	// directly - same assertion as before, now against a real Docker 404
	// instead of a hand-set fake error.
	containerService, _, _ := newRealUpgradeFixture(t, testUpgradeEnv)

	name := test_helper.UniqueName(t, testUpgradeSvcName)

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: name, Image: testUpgradeImage},
	}

	j := &captureOldContainerJob{containerService: containerService, upgradeReq: payload, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when InspectSmerd fails")
	}
}

func TestCaptureOldContainerJob_Success(t *testing.T) {
	t.Parallel()

	// environments.DefaultEnvironmentName seeded with an empty suffix means
	// the container's real Docker name is the bare virtual name, so
	// containerService.InspectSmerd(ctx, name) succeeds directly - no
	// runtimes fallback needed for this test (see
	// TestCaptureOldContainerJob_SuffixAwareLookup for that path).
	containerService, runtimes, cli := newRealUpgradeFixture(t, environments.DefaultEnvironmentName)

	test_helper.EnsurePulled(t, cli, test_helper.HelloWorldAppImage)

	networkName := test_helper.UniqueName(t, testNetworkName)

	err := dockerutils.CreateNetwork(context.Background(), cli, networkName)
	if err != nil {
		t.Fatalf("unexpected error creating network: %v", err)
	}

	pm := realPortManager(t)

	hostPort, err := pm.GetPort()
	if err != nil {
		t.Fatalf("unexpected error getting a host port: %v", err)
	}

	rt, err := runtimes.Runtime(context.Background(), environments.DefaultEnvironmentName)
	if err != nil {
		t.Fatalf("unexpected error resolving runtime: %v", err)
	}

	name := test_helper.UniqueName(t, testUpgradeSvcName)
	volumeName := test_helper.UniqueName(t, "vol")

	createReq := container_runtime.ContainerCreateRequest{
		Config: &container_runtime.ContainerConfig{Config: &container.Config{
			Image:  test_helper.HelloWorldAppImage,
			Env:    []string{testEnvKeyFoo + "=" + testEnvFoo},
			Labels: map[string]string{"team": "core"},
		}},
		HostConfig: &container_runtime.HostConfig{HostConfig: &container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port(testUpgradePgPort): []nat.PortBinding{
					{HostIP: "0.0.0.0", HostPort: strconv.FormatUint(uint64(hostPort), 10)},
				},
			},
			Mounts: []mount.Mount{{Type: mount.TypeVolume, Source: volumeName, Target: "/data"}},
		}},
		ContainerName: name,
	}

	created, err := rt.ContainerCreate(context.Background(), createReq)
	if err != nil {
		t.Fatalf("unexpected error creating container: %v", err)
	}

	t.Cleanup(func() { test_helper.RemoveContainer(t, cli, created.ID) })

	connReq := dockerutils.ConnectToNetworkRequest{
		NetworkName: networkName,
		ContId:      created.ID,
		Aliases:     []string{testNetworkAlias},
	}

	err = dockerutils.ConnectToNetwork(context.Background(), cli, connReq)
	if err != nil {
		t.Fatalf("unexpected error connecting to network: %v", err)
	}

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: name, Image: testUpgradeImage},
	}

	j := &captureOldContainerJob{containerService: containerService, upgradeReq: payload, ctx: payload}

	err = j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetOldContainerId() != created.ID {
		t.Errorf("expected old container id %q, got %q", created.ID, payload.GetOldContainerId())
	}

	req := payload.GetRequest()
	if req.GetName() != name {
		t.Errorf("expected request name %q, got %q", name, req.GetName())
	}

	if req.GetImageName() != testUpgradeImage {
		t.Errorf("expected request image %q (the upgrade target, not the old image), got %q",
			testUpgradeImage, req.GetImageName())
	}

	if req.GetEnv()[testEnvKeyFoo] != testEnvFoo {
		t.Errorf("expected env carried over from old container, got %v", req.GetEnv())
	}

	// Bonus real-Docker coverage vs. the old hand-simulated fake: Docker
	// itself adds the container's own short ID as a network alias, so this
	// now exercises fromContainerNetwork's self-uuid filtering against a
	// genuine Docker-populated alias list rather than a fixture value that
	// merely mimicked one.
	if len(req.GetSettings().GetNetwork()) != 1 || len(req.GetSettings().GetNetwork()[0].GetAliases()) != 1 {
		t.Errorf("expected the container's own id filtered out of network aliases, got %v",
			req.GetSettings().GetNetwork())
	}
}

// RED (runtime failure, compiles fine). captureOldContainerJob.Do builds a
// CreateSmerd_Request without ever copying Environment onto it, even though
// every downstream step in the upgrade pipeline (port locking via
// prepareUpgradeVervConfigJob.lockPorts, network creation, etc.) reads
// Environment off that same request. This is root cause #2 of the
// suffixed-environment upgrade bug (see tests/e2e's
// Test_UpgradeSmerd_InSuffixedEnvironment for the end-to-end repro).
func TestCaptureOldContainerJob_SetsEnvironmentFromUpgradeRequest(t *testing.T) {
	t.Parallel()

	// Cheapest real fixture: one container, no ports/volumes/network, real
	// ContainerCreate - created through testUpgradeEnv's resolved runtime (so
	// its real Docker name carries that environment's suffix), matching the
	// payload's own Environment below. containerService.InspectSmerd is now
	// environment-scoped (see container_manager.InspectSmerd), so the old
	// container must actually live in the environment the request names for
	// the direct, no-fallback lookup below to succeed (j.runtimes stays nil -
	// this test isn't exercising the fallback path, see
	// TestCaptureOldContainerJob_SuffixAwareLookup for that).
	containerService, runtimes, cli := newRealUpgradeFixture(t, testUpgradeEnv)

	test_helper.EnsurePulled(t, cli, test_helper.HelloWorldAppImage)

	rt, err := runtimes.Runtime(context.Background(), testUpgradeEnv)
	if err != nil {
		t.Fatalf("unexpected error resolving runtime: %v", err)
	}

	name := test_helper.UniqueName(t, testUpgradeSvcName)

	createReq := container_runtime.ContainerCreateRequest{
		Config:        &container_runtime.ContainerConfig{Config: &container.Config{Image: test_helper.HelloWorldAppImage}},
		ContainerName: name,
	}

	created, err := rt.ContainerCreate(context.Background(), createReq)
	if err != nil {
		t.Fatalf("unexpected error creating container: %v", err)
	}

	t.Cleanup(func() { test_helper.RemoveContainer(t, cli, created.ID) })

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{
			Name:        name,
			Image:       testUpgradeImage,
			Environment: testUpgradeEnv,
		},
	}

	j := &captureOldContainerJob{containerService: containerService, upgradeReq: payload, ctx: payload}

	err = j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := payload.GetRequest()
	if req.GetEnvironment() != testUpgradeEnv {
		t.Errorf("expected request Environment %q (mirroring the upgrade request's), got %q",
			testUpgradeEnv, req.GetEnvironment())
	}
}

// RED (compile-time). captureOldContainerJob resolves "the current
// container" via containerService.InspectSmerd(ctx, bareName) alone - a raw,
// suffix-blind lookup. In a real suffixed environment the actual Docker
// container is named "name_<suffix>"
// (container_runtime.labelBasedRuntime.containerName), so a bare-name
// InspectSmerd 404s before any upgrade logic runs - this is root cause #3 of
// the suffixed-environment upgrade bug.
//
// The fix needs a `runtimes container_runtime.RuntimeResolver` field on
// captureOldContainerJob so it can fall back to a suffix-aware resolution
// (the same resolveOwnedContainer-style lookup labelBasedRuntime.Remove
// already does) instead of trusting containerService.InspectSmerd to already
// be suffix-aware.
//
// This test simulates today's real-world failure - containerService's
// bare-name InspectSmerd 404s, exactly like the real Docker daemon does in a
// suffixed environment - and asserts Do() must still succeed by falling back
// to the runtimes resolver: resolving a real environments.NewStatic-backed
// environment, listing containers through it (fakeContainerRuntime.ListContainers,
// which is what the fallback actually calls), and re-inspecting by the
// resolved id. inspectFailFor is set to the bare name only, so a call with any
// other id (the resolved one) succeeds - if the fallback were skipped or
// masked a real resolution error instead of surfacing it, this would fail for
// a different reason than "not implemented yet", proving the fallback path
// itself is what's under test, not just "doesn't error".
// TestCaptureOldContainerJob_SuffixAwareLookup is the highest-value
// conversion in this file: the fixture container is created THROUGH the
// resolved runtime (so its real Docker name is genuinely suffixed,
// "<name>_STAGE" - see environments.NewStatic's "named environment's suffix
// defaults to its own name" rule), then Do is called with the bare virtual
// name. containerService.InspectSmerd(ctx, bareName) gets a real 404 from
// the daemon (the real container is named "<name>_STAGE", not "<name>"), so
// this only passes if resolveCurrentContainer's real fallback - ListContainers
// through the resolved runtime, then re-inspect by the resolved id - actually
// runs, directly re-proving the bug class fixed in 6b5f97c2/7fcb6079 against
// a real daemon instead of a hand-simulated one.
func TestCaptureOldContainerJob_SuffixAwareLookup(t *testing.T) {
	t.Parallel()

	containerService, runtimes, cli := newRealUpgradeFixture(t, testUpgradeEnv)

	test_helper.EnsurePulled(t, cli, test_helper.HelloWorldAppImage)

	name := test_helper.UniqueName(t, testUpgradeSvcName)

	rt, err := runtimes.Runtime(context.Background(), testUpgradeEnv)
	if err != nil {
		t.Fatalf("unexpected error resolving runtime: %v", err)
	}

	createReq := container_runtime.ContainerCreateRequest{
		Config:        &container_runtime.ContainerConfig{Config: &container.Config{Image: test_helper.HelloWorldAppImage}},
		ContainerName: name,
	}

	created, err := rt.ContainerCreate(context.Background(), createReq)
	if err != nil {
		t.Fatalf("unexpected error creating container: %v", err)
	}

	t.Cleanup(func() { test_helper.RemoveContainer(t, cli, created.ID) })

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{
			Name:        name,
			Image:       testUpgradeImage,
			Environment: testUpgradeEnv,
		},
	}

	j := &captureOldContainerJob{
		containerService: containerService,
		upgradeReq:       payload,
		ctx:              payload,
		runtimes:         runtimes,
	}

	err = j.Do(context.Background())
	if err != nil {
		t.Fatalf(
			"expected a suffix-aware lookup via runtimes to succeed despite a bare-name InspectSmerd 404, got: %v",
			err,
		)
	}

	if payload.GetRequest().GetName() != name {
		t.Errorf("expected the container found via the fallback to be captured, got name %q",
			payload.GetRequest().GetName())
	}

	if payload.GetOldContainerId() != created.ID {
		t.Errorf("expected old container id %q captured via the fallback, got %q",
			created.ID, payload.GetOldContainerId())
	}
}

// RED (compile-time). checkSelfUpgradeJob has the identical gap as
// captureOldContainerJob above: it inspects "the current container" via a
// bare-name, suffix-blind containerService.InspectSmerd call. It needs the
// same `runtimes container_runtime.RuntimeResolver` field.
//
// Note this test cannot exercise the self-upgrade branch itself (see
// TestCheckSelfUpgradeJob_NotInsideContainer_NoOp's comment: env.GetContainerId()
// always returns nil under `go test`), so it only documents/proves the
// missing field via a compile failure - not a behavioral runtime assertion.
func TestCheckSelfUpgradeJob_HasRuntimesFieldForSuffixAwareLookup(t *testing.T) {
	t.Parallel()

	containerService, runtimes, _ := newRealUpgradeFixture(t, testUpgradeEnv)

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: testUpgradeSvcName, Environment: testUpgradeEnv},
	}

	j := &checkSelfUpgradeJob{
		containerService: containerService,
		upgradeReq:       payload,
		runtimes:         runtimes,
	}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// prepareUpgradeImageJob

func TestPrepareUpgradeImageJob_PullError(t *testing.T) {
	docker := newFakeDocker()

	docker.pullImageErr = errNoSuchImage

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Image: testUpgradeImage},
	}

	j := &prepareUpgradeImageJob{runtimes: newFakeRuntimes(docker, nil), upgradeReq: payload, ctx: payload}

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

	j := &prepareUpgradeImageJob{runtimes: newFakeRuntimes(docker, nil), upgradeReq: payload, ctx: payload}

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

// TestPauseOldContainerJob_Success is exercised against a real local Docker
// daemon (newRealUpgradeFixture), not a hand-written fake, per this repo's
// testing policy (CLAUDE.md's "Testing" section) - Do/Rollback are now
// touching this test's own code (networkBindingsFor,
// ContainerRuntime.DisconnectFromNetworks/ConnectToNetwork), so the fake this
// test used to lean on (fakeContainerAPI's NetworkDisconnect/NetworkConnect)
// is retired rather than kept alongside a parallel real-Docker path.
func TestPauseOldContainerJob_Success(t *testing.T) {
	// Not t.Parallel(): kept sequential relative to
	// TestPrepareUpgradeVervConfigJob_Success (both create real networks on
	// the shared daemon); dockerutils.CreateNetwork no longer hardcodes a
	// subnet (see network.go), but there's no need to prove out parallel
	// safety here when this test's actual subject is pause/rollback, not
	// network-creation concurrency.
	_, runtimes, cli := newRealUpgradeFixture(t, testUpgradeEnv)

	test_helper.EnsurePulled(t, cli, test_helper.HelloWorldAppImage)

	name := test_helper.UniqueName(t, testUpgradeSvcName)

	runtime, err := runtimes.Runtime(context.Background(), testUpgradeEnv)
	require.NoError(t, err)

	hostCfg := &container.HostConfig{
		PortBindings: nat.PortMap{testUpgradePgPort: []nat.PortBinding{{HostPort: "40001"}}},
	}

	createReq := container_runtime.ContainerCreateRequest{
		Config: &container_runtime.ContainerConfig{Config: &container.Config{
			Image:        test_helper.HelloWorldAppImage,
			ExposedPorts: nat.PortSet{testUpgradePgPort: struct{}{}},
		}},
		HostConfig:    &container_runtime.HostConfig{HostConfig: hostCfg},
		ContainerName: name,
	}

	created, err := runtime.ContainerCreate(context.Background(), createReq)
	require.NoError(t, err)

	err = cli.ContainerStart(context.Background(), created.ID, container.StartOptions{})
	require.NoError(t, err)

	err = runtime.CreateNetwork(context.Background(), env.VervNetwork)
	require.NoError(t, err)

	vervNetName := env.VervNetwork + "_" + testUpgradeEnv

	// Registration order matters: t.Cleanup runs LIFO, and by the end of this
	// test (after job.Rollback reconnects the container - see below) the
	// container is attached to vervNetName again, so removing the network
	// must happen AFTER the container is gone, not before (Docker refuses to
	// remove a network with an active endpoint - silently, since NetworkRemove's
	// error is ignored here - which would leak the network and collide with
	// any other test's identically-subnetted network, exactly the flake
	// TestCaptureOldContainerJob_Success is already known for).
	t.Cleanup(func() {
		_ = cli.NetworkRemove(context.Background(), vervNetName)
	})

	t.Cleanup(func() {
		test_helper.RemoveContainer(t, cli, created.ID)
	})

	connReq := container_runtime.ConnectToNetworkRequest{
		ContainerID: created.ID,
		NetworkName: env.VervNetwork,
		Aliases:     []string{name},
	}

	err = runtime.ConnectToNetwork(context.Background(), connReq)
	require.NoError(t, err)

	// Settings.Ports must be non-empty here to match the real container: it was
	// created (above) with a host port binding, and networkBindingsFor only
	// includes the default network when the request says it has one - same
	// gating createContainerJob.connectNetworks applies at create time (see
	// networkBindingsFor's doc comment).
	payload := &velez_api.UpgradeSmerdTaskPayload{
		OldContainerId: strPtr(created.ID),
		Request: &velez_api.CreateSmerd_Request{
			Name:        name,
			Environment: testUpgradeEnv,
			Settings: &velez_api.Container_Settings{
				Ports: []*velez_api.Port{{ServicePortNumber: 5432}},
			},
		},
	}

	job := &pauseOldContainerJob{
		dockerAPI:   cli,
		portManager: realPortManager(t),
		runtimes:    runtimes,
		req:         payload,
		ctx:         payload,
	}

	err = job.Do(context.Background())
	require.NoError(t, err)

	inspected, err := cli.ContainerInspect(context.Background(), created.ID)
	require.NoError(t, err)
	require.NotContains(t, inspected.NetworkSettings.Networks, vervNetName,
		"expected the container to be disconnected from its network")

	// Real Docker reports one binding per IP family for the same host port
	// when no HostIP is pinned (IPv4 + IPv6), so portsOnHold can legitimately
	// hold 40001 more than once - unlike the hand-crafted single-binding
	// fixture this test used to use. What matters is that every entry is the
	// port actually published, not the exact count.
	require.NotEmpty(t, job.portsOnHold)

	for _, p := range job.portsOnHold {
		require.Equal(t, uint32(40001), p)
	}

	err = job.Rollback(context.Background())
	require.NoError(t, err)

	inspected, err = cli.ContainerInspect(context.Background(), created.ID)
	require.NoError(t, err)
	require.Contains(t, inspected.NetworkSettings.Networks, vervNetName,
		"expected the container to be reconnected to its network on rollback")
	require.Contains(t, inspected.NetworkSettings.Networks[vervNetName].Aliases, name,
		"expected the reconnect to carry the same alias the container had before pause")
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
		runtimes:    newFakeRuntimes(docker, nil),
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

// TestPrepareUpgradeVervConfigJob_Success is exercised against a real local
// Docker daemon (newRealUpgradeFixture), not a hand-written fake, per this
// repo's testing policy (CLAUDE.md's "Testing" section) - Do now creates the
// request's extra networks through the resolved ContainerRuntime instead of
// against a raw client.APIClient, so the fake this test used to lean on
// (fakeContainerAPI's NetworkList/NetworkCreate) is retired.
func TestPrepareUpgradeVervConfigJob_Success(t *testing.T) {
	// Not t.Parallel(): kept sequential relative to
	// TestPauseOldContainerJob_Success - see its identical note.
	_, runtimes, cli := newRealUpgradeFixture(t, testUpgradeEnv)

	networkName := test_helper.UniqueName(t, testNetworkName)

	payload := &velez_api.UpgradeSmerdTaskPayload{
		Request: &velez_api.CreateSmerd_Request{
			Name:        testUpgradeSvcName,
			Environment: testUpgradeEnv,
			Env:         map[string]string{},
			Labels:      map[string]string{},
			Settings: &velez_api.Container_Settings{
				Ports:   []*velez_api.Port{{ServicePortNumber: 8080}},
				Network: []*velez_api.NetworkBind{{NetworkName: networkName}},
			},
		},
		ImageLabels: map[string]string{"custom": "label"},
	}

	pm := realPortManager(t)

	j := &prepareUpgradeVervConfigJob{portManager: pm, runtimes: runtimes, imageMeta: payload, req: payload}

	err := j.Do(context.Background())
	require.NoError(t, err)

	require.NotZero(t, payload.GetRequest().GetSettings().GetPorts()[0].GetExposedTo(),
		"expected a host port to be locked")
	require.Equal(t, "label", payload.GetRequest().GetLabels()["custom"],
		"expected image labels merged into request labels")
	require.Equal(t, testUpgradeSvcName, payload.GetRequest().GetLabels()[labels.ComposeGroupLabel])

	expectedNetName := networkName + "_" + testUpgradeEnv

	t.Cleanup(func() {
		_ = cli.NetworkRemove(context.Background(), expectedNetName)
	})

	_, err = cli.NetworkInspect(context.Background(), expectedNetName, network.InspectOptions{})
	require.NoError(t, err, "expected the extra network to be created, suffixed to the environment")

	err = j.Rollback(context.Background())
	require.NoError(t, err)
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

// renameContainerJob - runtime-scoped rename
//
// TestRenameContainerJob_SuccessAndRollback previously exercised
// renameContainerJob against a raw dockerAPI (ContainerInspect capturing
// oldName at Do-time, then ContainerRename directly) - that shape is gone
// now that renameContainerJob resolves a ContainerRuntime for the request's
// environment and calls its Rename with a precomputed virtual oldName/newName
// (see upgradeSmerdHandler.BuildJobs). The two tests below supersede it,
// covering the same Do-then-Rollback round trip through the new shape.
//
// renameContainerJob.Do/Rollback used to call the raw dockerAPI.ContainerRename
// directly with a bare, un-suffixed name - so in a suffixed environment the
// container's REAL Docker name never carried the environment's suffix (e.g.
// it ended up literally "mysvc_old" instead of "mysvc_old_STAGE"). See
// tests/e2e/suite_upgrade_smerd_test.go's Test_UpgradeSmerd_InSuffixedEnvironment
// for the e2e proof (it inspects the real Docker daemon directly).
func TestRenameContainerJob_Do_ResolvesRuntimeAndRenamesViaRuntime(t *testing.T) {
	t.Parallel()

	_, runtimes, cli := newRealUpgradeFixture(t, testUpgradeEnv)

	test_helper.EnsurePulled(t, cli, test_helper.HelloWorldAppImage)

	name := test_helper.UniqueName(t, testUpgradeSvcName)
	newName := test_helper.UniqueName(t, "newname")

	rt, err := runtimes.Runtime(context.Background(), testUpgradeEnv)
	if err != nil {
		t.Fatalf("unexpected error resolving runtime: %v", err)
	}

	createReq := container_runtime.ContainerCreateRequest{
		Config:        &container_runtime.ContainerConfig{Config: &container.Config{Image: test_helper.HelloWorldAppImage}},
		ContainerName: name,
	}

	created, err := rt.ContainerCreate(context.Background(), createReq)
	if err != nil {
		t.Fatalf("unexpected error creating container: %v", err)
	}

	t.Cleanup(func() { test_helper.RemoveContainer(t, cli, created.ID) })

	payload := &velez_api.UpgradeSmerdTaskPayload{
		ContainerId:    strPtr(created.ID),
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: name, Environment: testUpgradeEnv},
	}

	j := &renameContainerJob{
		runtimes: runtimes,
		req:      payload,
		ctx:      payload,
		newName:  newName,
	}

	err = j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// A non-default named environment's suffix defaults to its own name
	// (environments.NewStatic's seeding rule), so the renamed container's
	// real Docker name is newName + "_" + testUpgradeEnv.
	suffixedNewName := newName + "_" + testUpgradeEnv

	inspected, err := cli.ContainerInspect(context.Background(), suffixedNewName)
	if err != nil {
		t.Fatalf("expected the renamed, suffixed container %q to exist on the daemon: %v", suffixedNewName, err)
	}

	if inspected.ID != created.ID {
		t.Errorf("expected the renamed container's id to be unchanged, got %q want %q", inspected.ID, created.ID)
	}
}

func TestRenameContainerJob_Rollback_RenamesBackToOldNameViaRuntime(t *testing.T) {
	t.Parallel()

	_, runtimes, cli := newRealUpgradeFixture(t, testUpgradeEnv)

	test_helper.EnsurePulled(t, cli, test_helper.HelloWorldAppImage)

	name := test_helper.UniqueName(t, testUpgradeSvcName)
	oldName := test_helper.UniqueName(t, testUpgradeSvcName+oldContainerSuffix)

	rt, err := runtimes.Runtime(context.Background(), testUpgradeEnv)
	if err != nil {
		t.Fatalf("unexpected error resolving runtime: %v", err)
	}

	createReq := container_runtime.ContainerCreateRequest{
		Config:        &container_runtime.ContainerConfig{Config: &container.Config{Image: test_helper.HelloWorldAppImage}},
		ContainerName: name,
	}

	created, err := rt.ContainerCreate(context.Background(), createReq)
	if err != nil {
		t.Fatalf("unexpected error creating container: %v", err)
	}

	t.Cleanup(func() { test_helper.RemoveContainer(t, cli, created.ID) })

	payload := &velez_api.UpgradeSmerdTaskPayload{
		ContainerId:    strPtr(created.ID),
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: name, Environment: testUpgradeEnv},
	}

	j := &renameContainerJob{
		runtimes: runtimes,
		req:      payload,
		ctx:      payload,
		newName:  name,
		oldName:  oldName,
	}

	err = j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected rollback error: %v", err)
	}

	suffixedOldName := oldName + "_" + testUpgradeEnv

	inspected, err := cli.ContainerInspect(context.Background(), suffixedOldName)
	if err != nil {
		t.Fatalf("expected the rolled-back, suffixed container %q to exist on the daemon: %v", suffixedOldName, err)
	}

	if inspected.ID != created.ID {
		t.Errorf("expected the rolled-back container's id to be unchanged, got %q want %q", inspected.ID, created.ID)
	}
}

// dropOwnedContainerJob - Stage B (runtime-scoped drop)
//
// RED at COMPILE time. dropOwnedContainerJob does not exist yet. It replaces
// dropScratchContainerJob at upgrade_smerd.go's two drop call sites
// (stepDropConfigFetcherContainer, stepDropOldContainer) - both currently call
// node_clients.Docker.Remove directly, bypassing the environment-scoped
// ContainerRuntime entirely, same class of bug dropContainerJob (drop_smerd.go)
// already fixed for DropSmerd. See that type's doc comment for the identical
// rationale (idempotent removal, environment-scoped).
func TestDropOwnedContainerJob_Do_ResolvesRuntimeAndRemoves(t *testing.T) {
	t.Parallel()

	_, runtimes, cli := newRealUpgradeFixture(t, testUpgradeEnv)

	test_helper.EnsurePulled(t, cli, test_helper.HelloWorldAppImage)

	name := test_helper.UniqueName(t, testUpgradeSvcName)

	rt, err := runtimes.Runtime(context.Background(), testUpgradeEnv)
	if err != nil {
		t.Fatalf("unexpected error resolving runtime: %v", err)
	}

	createReq := container_runtime.ContainerCreateRequest{
		Config:        &container_runtime.ContainerConfig{Config: &container.Config{Image: test_helper.HelloWorldAppImage}},
		ContainerName: name,
	}

	created, err := rt.ContainerCreate(context.Background(), createReq)
	if err != nil {
		t.Fatalf("unexpected error creating container: %v", err)
	}

	// Defensive removal regardless of Do's own outcome - belt-and-suspenders,
	// mirrors tests/e2e's clean().
	t.Cleanup(func() { test_helper.RemoveContainer(t, cli, created.ID) })

	// dropOwnedContainerJob resolves its runtime off req.GetRequest() (a
	// smerdRequestAccessor), not UpgradeRequest - Environment must be set
	// here for the resolved runtime's suffix to match the one the fixture
	// container was created under.
	payload := &velez_api.UpgradeSmerdTaskPayload{
		ContainerId: strPtr(created.ID),
		Request:     &velez_api.CreateSmerd_Request{Name: name, Environment: testUpgradeEnv},
	}

	j := &dropOwnedContainerJob{runtimes: runtimes, req: payload, ctx: payload}

	err = j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = cli.ContainerInspect(context.Background(), created.ID)
	if err == nil {
		t.Fatalf("expected container %q to be removed, but ContainerInspect succeeded", created.ID)
	}
}

func TestDropOwnedContainerJob_Do_EmptyContainerId_NoOp(t *testing.T) {
	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: testUpgradeSvcName, Environment: testUpgradeEnv},
	}

	j := &dropOwnedContainerJob{req: payload, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error for an empty container id: %v", err)
	}
}

// createContainerJob.Rollback / renamingCreateContainerJob.Rollback -
// runtime-scoped removal (bonus coverage, per the jobs cutover plan)
//
// These two are RED at RUNTIME (not compile time - createContainerJob already
// has req/runtimes fields, wired for Do but never read by Rollback today).
// createContainerJob.Rollback removes the just-created container via
// j.nodeClients.Docker().Remove directly, completely bypassing the
// environment-scoped ContainerRuntime resolved via j.runtimes for j.req's
// environment. Each test below wires TWO separate fake dockers - one behind
// nodeClients (which Rollback must NOT use), one behind runtimes (which it
// must use instead) - specifically so "which path got called" is
// observable; a single shared fake docker behind both would make the two
// paths indistinguishable.
func TestCreateContainerJob_Rollback_RemovesViaResolvedRuntime(t *testing.T) {
	directDocker := newFakeDocker()
	scopedDocker := newFakeDocker()
	runtimes := newFakeRuntimes(scopedDocker, environments.NewStatic([]string{testUpgradeEnv}, ""))

	payload := &velez_api.UpgradeSmerdTaskPayload{
		ContainerId: strPtr(testCreatedID),
		Request:     &velez_api.CreateSmerd_Request{Name: testUpgradeSvcName, Environment: testUpgradeEnv},
	}

	j := &createContainerJob{
		nodeClients: newFakeNodeClients(directDocker),
		req:         payload,
		ctx:         payload,
		runtimes:    runtimes,
	}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected rollback error: %v", err)
	}

	if len(directDocker.removeCalledWith) != 0 {
		t.Errorf("expected Rollback NOT to call nodeClients.Docker().Remove directly, got %v",
			directDocker.removeCalledWith)
	}

	if len(scopedDocker.removeCalledWith) != 1 || scopedDocker.removeCalledWith[0] != testCreatedID {
		t.Errorf("expected Rollback to remove %q through the runtime resolved for environment %q, got %v",
			testCreatedID, testUpgradeEnv, scopedDocker.removeCalledWith)
	}
}

// renamingCreateContainerJob.Rollback constructs its inner createContainerJob
// as `&createContainerJob{nodeClients: j.nodeClients, ctx: j.ctx}` today -
// deliberately (if silently) omitting req/runtimes, harmless only because
// Rollback never reads them yet. The moment createContainerJob.Rollback is
// fixed to resolve via j.runtimes for j.req's environment (the test above),
// this omission becomes a nil-deref waiting to happen. This test proves the
// propagation gap directly: it must currently fail the same way the test
// above does, for the same reason (the inner job never got req/runtimes, so
// it falls back to nodeClients - here, the "wrong" docker).
func TestRenamingCreateContainerJob_Rollback_PropagatesReqAndRuntimesToInnerJob(t *testing.T) {
	directDocker := newFakeDocker()
	scopedDocker := newFakeDocker()
	runtimes := newFakeRuntimes(scopedDocker, environments.NewStatic([]string{testUpgradeEnv}, ""))

	payload := &velez_api.UpgradeSmerdTaskPayload{
		ContainerId: strPtr(testCreatedID),
		Request:     &velez_api.CreateSmerd_Request{Name: testUpgradeSvcName, Environment: testUpgradeEnv},
	}

	j := &renamingCreateContainerJob{
		nodeClients: newFakeNodeClients(directDocker),
		req:         payload,
		ctx:         payload,
		runtimes:    runtimes,
		newName:     func(current string) string { return current },
	}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected rollback error: %v", err)
	}

	if len(scopedDocker.removeCalledWith) != 1 || scopedDocker.removeCalledWith[0] != testCreatedID {
		t.Errorf("expected the inner createContainerJob to carry req/runtimes through to Rollback and remove "+
			"%q via the resolved runtime, got scopedDocker=%v directDocker=%v",
			testCreatedID, scopedDocker.removeCalledWith, directDocker.removeCalledWith)
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

	handler := NewUpgradeSmerdHandler(nodeClients, containerService, configService, newFakeRuntimes(docker, nil))

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

	runtimes := newFakeRuntimes(docker, nil)

	runtimes.createNetworkErr = errNetworkCreate

	handler := NewUpgradeSmerdHandler(
		nodeClients, containerService, newFakeConfigurationService(), runtimes)

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
