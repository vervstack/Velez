package jobs

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

const (
	testEnvStage   = "STAGE"
	testEnvSvcName = "svc"
	testProdSuffix = "prod"
)

// envProvider is the minimal EnvironmentsProvider a job needs.
type envProvider struct {
	storage storage.EnvironmentsStorage
}

func (p *envProvider) Environments() storage.EnvironmentsStorage {
	return p.storage
}

func newEnvProvider(defaultSuffix string) *envProvider {
	return &envProvider{storage: newEnvStorage(defaultSuffix)}
}

// newEnvStorage builds the in-memory environments storage both the provider
// above and the fake runtime resolver (fakes_test.go) resolve against.
func newEnvStorage(defaultSuffix string) storage.EnvironmentsStorage {
	return environments.NewStatic(nil, defaultSuffix)
}

// --- TaskContext round-trip -------------------------------------------------
//
// internal/jobs persists TaskContext via plain encoding/json (checkpoint.go /
// worker.go), so every field a job later reads back must survive a
// marshal -> unmarshal cycle. CreateSmerd.Request's `config` oneof - the
// historical hazard called out in CLAUDE.md - was removed at the proto level
// (see create_smerd_request_roundtrip_test.go), but the new `environment`
// field is what createContainerJob now resolves its Docker suffix from, so it
// gets its own proof rather than an assumption.

func TestCreateSmerdTaskPayload_EnvironmentSurvivesRoundTrip(t *testing.T) {
	req := &velez_api.CreateSmerd_Request{Name: testEnvSvcName, Environment: testEnvStage}

	payload := &velez_api.CreateSmerdTaskPayload{}
	payload.SetRequest(req)

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	restored := &velez_api.CreateSmerdTaskPayload{}

	err = json.Unmarshal(data, restored)
	require.NoError(t, err)

	require.Equal(t, testEnvStage, restored.GetRequest().GetEnvironment())
	require.Equal(t, testEnvSvcName, restored.GetRequest().GetName())
}

func TestDropSmerdTaskPayload_EnvironmentSurvivesRoundTrip(t *testing.T) {
	payload := &velez_api.DropSmerdTaskPayload{}
	payload.SetRequest(&velez_api.DropSmerd_Request{Name: []string{testEnvSvcName}, Environment: testEnvStage})

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	restored := &velez_api.DropSmerdTaskPayload{}
	require.NoError(t, json.Unmarshal(data, restored))
	require.Equal(t, testEnvStage, restored.GetRequest().GetEnvironment())
}

func TestUpgradeSmerdTaskPayload_EnvironmentSurvivesRoundTrip(t *testing.T) {
	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: testEnvSvcName, Environment: testEnvStage},
		Request:        &velez_api.CreateSmerd_Request{Name: testEnvSvcName, Environment: testEnvStage},
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	restored := &velez_api.UpgradeSmerdTaskPayload{}
	require.NoError(t, json.Unmarshal(data, restored))
	require.Equal(t, testEnvStage, restored.GetUpgradeRequest().GetEnvironment())
	require.Equal(t, testEnvStage, restored.GetRequest().GetEnvironment())
}

// --- resolveEnvironmentSuffix ----------------------------------------------

func TestResolveEnvironmentSuffix_ResolvesKnownEnvironment(t *testing.T) {
	provider := newEnvProvider(testProdSuffix)

	suffix, err := resolveEnvironmentSuffix(
		context.Background(), provider, environments.DefaultEnvironmentName)
	require.NoError(t, err)
	require.Equal(t, testProdSuffix, suffix)
}

// An empty name means "the default environment", so it picks up the default
// environment's suffix rather than staying unscoped. This is what keeps
// pre-environments callers (the UI, e2e tests, internally built requests)
// landing on the same containers they always did.
func TestResolveEnvironmentSuffix_EmptyNameResolvesDefaultEnvironment(t *testing.T) {
	suffix, err := resolveEnvironmentSuffix(context.Background(), newEnvProvider(testProdSuffix), "")
	require.NoError(t, err)
	require.Equal(t, testProdSuffix, suffix)
}

// Best-effort default resolution: with no environments storage at all (boot
// time, single-node before the storage swap) the default degrades to the empty
// suffix - exactly what an unconfigured ContainerSuffix produced - instead of
// failing the job.
func TestResolveEnvironmentSuffix_EmptyNameWithoutProviderIsUnscoped(t *testing.T) {
	suffix, err := resolveEnvironmentSuffix(context.Background(), nil, "")
	require.NoError(t, err)
	require.Empty(t, suffix)
}

// The default environment and an explicitly named one must not collide: each
// resolves to its own suffix.
func TestResolveEnvironmentSuffix_ExplicitEnvironmentGetsOwnSuffix(t *testing.T) {
	ctx := context.Background()
	provider := &envProvider{storage: environments.NewStatic([]string{testEnvStage}, testProdSuffix)}

	defaultSuffix, err := resolveEnvironmentSuffix(ctx, provider, "")
	require.NoError(t, err)

	stageSuffix, err := resolveEnvironmentSuffix(ctx, provider, testEnvStage)
	require.NoError(t, err)

	require.Equal(t, testProdSuffix, defaultSuffix)
	require.Equal(t, testEnvStage, stageSuffix)
	require.NotEqual(t, defaultSuffix, stageSuffix)
}

func TestResolveEnvironmentSuffix_UnknownEnvironmentErrors(t *testing.T) {
	_, err := resolveEnvironmentSuffix(context.Background(), newEnvProvider(""), "ghost")
	require.Error(t, err)
}

func TestResolveEnvironmentSuffix_NoProviderErrors(t *testing.T) {
	_, err := resolveEnvironmentSuffix(context.Background(), nil, testEnvStage)
	require.Error(t, err)
}

// --- createContainerJob threading ------------------------------------------

// The suffix stamped onto the container must come from the request's
// environment, resolved at run time - not from anything captured when the
// handler was built.
func TestCreateContainerJob_StampsResolvedEnvironmentSuffix(t *testing.T) {
	docker := newFakeDocker()
	// ContainerCreate records the suffix before failing, so the assertion
	// doesn't need a full client.APIClient fake for the ContainerInspect call
	// that would follow a successful create.
	docker.containerCreateErr = rerrors.New("stop here")

	nodeClients := newFakeNodeClients(docker)

	payload := &velez_api.CreateSmerdTaskPayload{}
	payload.SetRequest(&velez_api.CreateSmerd_Request{
		Name:        testEnvSvcName,
		Environment: environments.DefaultEnvironmentName,
	})

	job := &createContainerJob{
		nodeClients: nodeClients,
		req:         payload,
		ctx:         payload,
		runtimes:    newFakeRuntimes(docker, newEnvStorage(testProdSuffix)),
	}

	err := job.Do(context.Background())
	require.Error(t, err)
	require.Equal(t, []string{testProdSuffix}, docker.containerCreateSuffixes)
}

// The backward-compatibility case: a CreateSmerd request that carries no
// environment at all (every caller that predates the feature) must still be
// stamped with the DEFAULT environment's suffix, not an empty one.
func TestCreateContainerJob_EmptyEnvironmentStampsDefaultSuffix(t *testing.T) {
	docker := newFakeDocker()

	docker.containerCreateErr = rerrors.New("stop here")

	nodeClients := newFakeNodeClients(docker)

	payload := &velez_api.CreateSmerdTaskPayload{}
	payload.SetRequest(&velez_api.CreateSmerd_Request{Name: testEnvSvcName})

	job := &createContainerJob{
		nodeClients: nodeClients,
		req:         payload,
		ctx:         payload,
		runtimes:    newFakeRuntimes(docker, newEnvStorage(testProdSuffix)),
	}

	err := job.Do(context.Background())
	require.Error(t, err)
	require.Equal(t, []string{testProdSuffix}, docker.containerCreateSuffixes)
}

// Two containers with the same logical name in two different environments must
// be stamped with two different suffixes - that's what makes them distinct
// Docker resources.
func TestCreateContainerJob_SameNameInTwoEnvironmentsGetsDistinctSuffixes(t *testing.T) {
	docker := newFakeDocker()

	docker.containerCreateErr = rerrors.New("stop here")

	nodeClients := newFakeNodeClients(docker)
	runtimes := newFakeRuntimes(docker, environments.NewStatic([]string{testEnvStage}, testProdSuffix))

	for _, environment := range []string{"", testEnvStage} {
		payload := &velez_api.CreateSmerdTaskPayload{}
		payload.SetRequest(&velez_api.CreateSmerd_Request{
			Name:        testEnvSvcName,
			Environment: environment,
		})

		job := &createContainerJob{
			nodeClients: nodeClients,
			req:         payload,
			ctx:         payload,
			runtimes:    runtimes,
		}

		err := job.Do(context.Background())
		require.Error(t, err)
	}

	require.Equal(t, []string{testProdSuffix, testEnvStage}, docker.containerCreateSuffixes)
}

func TestCreateContainerJob_UnknownEnvironmentFailsJob(t *testing.T) {
	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)

	payload := &velez_api.CreateSmerdTaskPayload{}
	payload.SetRequest(&velez_api.CreateSmerd_Request{Name: testEnvSvcName, Environment: "ghost"})

	job := &createContainerJob{
		nodeClients: nodeClients,
		req:         payload,
		ctx:         payload,
		runtimes:    newFakeRuntimes(docker, newEnvStorage(testProdSuffix)),
	}

	err := job.Do(context.Background())
	require.Error(t, err)
	require.Empty(t, docker.containerCreateSuffixes, "container must not be created")
}
