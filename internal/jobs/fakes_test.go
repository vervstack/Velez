package jobs

import (
	"archive/tar"
	"bytes"
	"context"
	"database/sql"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.redsock.ru/evon"
	"go.vervstack.ru/makosh/pkg/makosh_be"
	"go.vervstack.ru/matreshka/pkg/matreshka"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/local_state"
	"go.vervstack.ru/Velez/internal/clients/node_clients/ports"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

// fakeTasksStorage and fakeJobsStorage are minimal in-memory implementations
// of storage.TasksStorage/storage.JobsStorage, following this repo's existing
// convention of hand-written test fakes (see verv_services/list_test.go)
// rather than generated mocks, for the small Querier interfaces here.

type fakeTasksStorage struct {
	mu     sync.Mutex
	nextID int64
	byID   map[int64]tasks_queries.VelezTask
}

func newFakeTasksStorage() *fakeTasksStorage {
	return &fakeTasksStorage{byID: make(map[int64]tasks_queries.VelezTask)}
}

func (f *fakeTasksStorage) CreateTask(_ context.Context, arg tasks_queries.CreateTaskParams) (tasks_queries.VelezTask, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, t := range f.byID {
		if t.EntityID == arg.EntityID && t.Action == arg.Action {
			return tasks_queries.VelezTask{}, sql.ErrNoRows
		}
	}

	f.nextID++
	task := tasks_queries.VelezTask{
		ID:        f.nextID,
		EntityID:  arg.EntityID,
		Action:    arg.Action,
		Status:    tasks_queries.VelezTaskStatusPENDING,
		Context:   arg.Context,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	f.byID[task.ID] = task

	return task, nil
}

func (f *fakeTasksStorage) GetTaskByEntityAction(_ context.Context, arg tasks_queries.GetTaskByEntityActionParams) (tasks_queries.VelezTask, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, t := range f.byID {
		if t.EntityID == arg.EntityID && t.Action == arg.Action {
			return t, nil
		}
	}

	return tasks_queries.VelezTask{}, sql.ErrNoRows
}

func (f *fakeTasksStorage) GetTaskById(_ context.Context, id int64) (tasks_queries.VelezTask, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	t, ok := f.byID[id]
	if !ok {
		return tasks_queries.VelezTask{}, sql.ErrNoRows
	}

	return t, nil
}

func (f *fakeTasksStorage) ClaimTask(_ context.Context, arg tasks_queries.ClaimTaskParams) (tasks_queries.VelezTask, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	staleThreshold := arg.ClaimedAt.Time

	for id, t := range f.byID {
		claimable := t.Status == tasks_queries.VelezTaskStatusPENDING ||
			(t.Status == tasks_queries.VelezTaskStatusRUNNING && t.ClaimedAt.Valid && t.ClaimedAt.Time.Before(staleThreshold))

		if !claimable {
			continue
		}

		t.Status = tasks_queries.VelezTaskStatusRUNNING
		t.ClaimedAt = sql.NullTime{Time: time.Now(), Valid: true}
		t.ClaimedBy = arg.ClaimedBy
		t.UpdatedAt = time.Now()
		f.byID[id] = t

		return t, nil
	}

	return tasks_queries.VelezTask{}, sql.ErrNoRows
}

func (f *fakeTasksStorage) UpdateTaskContext(_ context.Context, arg tasks_queries.UpdateTaskContextParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	t, ok := f.byID[arg.ID]
	if !ok {
		return sql.ErrNoRows
	}

	t.Context = arg.Context
	t.UpdatedAt = time.Now()
	f.byID[arg.ID] = t

	return nil
}

func (f *fakeTasksStorage) FinishTask(_ context.Context, arg tasks_queries.FinishTaskParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	t, ok := f.byID[arg.ID]
	if !ok {
		return sql.ErrNoRows
	}

	t.Status = arg.Status
	t.Error = arg.Error
	t.UpdatedAt = time.Now()
	f.byID[arg.ID] = t

	return nil
}

func (f *fakeTasksStorage) WithTx(_ *sql.Tx) *tasks_queries.Queries {
	return nil
}

func (f *fakeTasksStorage) get(id int64) tasks_queries.VelezTask {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.byID[id]
}

type fakeJobsStorage struct {
	mu   sync.Mutex
	rows map[string]jobs_queries.VelezJob
}

func newFakeJobsStorage() *fakeJobsStorage {
	return &fakeJobsStorage{rows: make(map[string]jobs_queries.VelezJob)}
}

func jobKey(taskID int64, jobName string) string {
	return strconv.FormatInt(taskID, 10) + "@" + jobName
}

func (f *fakeJobsStorage) GetJob(_ context.Context, arg jobs_queries.GetJobParams) (jobs_queries.VelezJob, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	row, ok := f.rows[jobKey(arg.TaskID, arg.JobName)]
	if !ok {
		return jobs_queries.VelezJob{}, sql.ErrNoRows
	}

	return row, nil
}

func (f *fakeJobsStorage) CreateRunningJob(_ context.Context, arg jobs_queries.CreateRunningJobParams) (jobs_queries.VelezJob, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := jobKey(arg.TaskID, arg.JobName)
	if _, ok := f.rows[key]; ok {
		return jobs_queries.VelezJob{}, sql.ErrNoRows
	}

	row := jobs_queries.VelezJob{
		TaskID:    arg.TaskID,
		JobName:   arg.JobName,
		Status:    jobs_queries.VelezJobStatusRUNNING,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	f.rows[key] = row

	return row, nil
}

func (f *fakeJobsStorage) FinishJob(_ context.Context, arg jobs_queries.FinishJobParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := jobKey(arg.TaskID, arg.JobName)
	row, ok := f.rows[key]
	if !ok {
		return sql.ErrNoRows
	}

	row.Status = arg.Status
	row.Error = arg.Error
	row.UpdatedAt = time.Now()
	f.rows[key] = row

	return nil
}

func (f *fakeJobsStorage) ListJobsByTask(_ context.Context, taskID int64) ([]jobs_queries.VelezJob, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	out := make([]jobs_queries.VelezJob, 0)
	for _, row := range f.rows {
		if row.TaskID == taskID {
			out = append(out, row)
		}
	}

	return out, nil
}

func (f *fakeJobsStorage) WithTx(_ *sql.Tx) *jobs_queries.Queries {
	return nil
}

func (f *fakeJobsStorage) seedDone(taskID int64, jobName string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.rows[jobKey(taskID, jobName)] = jobs_queries.VelezJob{
		TaskID:  taskID,
		JobName: jobName,
		Status:  jobs_queries.VelezJobStatusDONE,
	}
}

// fakeServicesStorage is a minimal in-memory implementation of
// storage.ServicesStorage for exercising create_service jobs without a
// database.
type fakeServicesStorage struct {
	mu       sync.Mutex
	upserted []string
}

func newFakeServicesStorage() *fakeServicesStorage {
	return &fakeServicesStorage{}
}

func (f *fakeServicesStorage) GetByName(_ context.Context, name string) (domain.Service, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, n := range f.upserted {
		if n == name {
			return domain.Service{ServiceBaseInfo: domain.ServiceBaseInfo{Name: name}}, nil
		}
	}

	return domain.Service{}, sql.ErrNoRows
}

func (f *fakeServicesStorage) UpsertService(_ context.Context, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.upserted = append(f.upserted, name)

	return nil
}

func (f *fakeServicesStorage) List(_ context.Context, _ domain.ListServicesReq) (domain.ServiceList, error) {
	return domain.ServiceList{}, nil
}

func (f *fakeServicesStorage) Delete(_ context.Context, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i, n := range f.upserted {
		if n == name {
			f.upserted = append(f.upserted[:i], f.upserted[i+1:]...)
			break
		}
	}

	return nil
}

// fakeDocker is a minimal in-memory implementation of node_clients.Docker
// for exercising assemble_config's Docker-touching jobs without a real
// Docker daemon. Only the methods those jobs actually call
// (PullImage/ContainerCreate/Remove) are configurable; the rest return zero
// values since the jobs under test never invoke them.
type fakeDocker struct {
	mu sync.Mutex

	pullImageResp image.InspectResponse
	pullImageErr  error

	containerCreateResp container.CreateResponse
	containerCreateErr  error

	removeErr        error
	removeCalledWith []string

	execResp       []byte
	execErr        error
	execCalledWith []container.ExecOptions

	// clientAPI, if set via withClient, is returned by Client() instead of
	// nil - lets tests that need the raw Docker engine client (e.g. jobs
	// depending on a narrow client.APIClient slice like pauseAPI/renameAPI)
	// inject a fakeContainerAPI instead of hitting a nil dereference.
	clientAPI client.APIClient
}

func newFakeDocker() *fakeDocker {
	return &fakeDocker{}
}

func (f *fakeDocker) PullImage(_ context.Context, _ string) (image.InspectResponse, error) {
	return f.pullImageResp, f.pullImageErr
}

func (f *fakeDocker) Remove(_ context.Context, uuid string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.removeCalledWith = append(f.removeCalledWith, uuid)

	return f.removeErr
}

func (f *fakeDocker) Stop(_ context.Context, _ string) error {
	return nil
}

func (f *fakeDocker) Restart(_ context.Context, _ string) error {
	return nil
}

func (f *fakeDocker) ListContainers(_ context.Context, _ *velez_api.ListSmerds_Request) ([]container.Summary, error) {
	return nil, nil
}

func (f *fakeDocker) ListOccupiedPorts(_ context.Context) ([]uint32, error) {
	return nil, nil
}

func (f *fakeDocker) Exec(_ context.Context, _ string, opts container.ExecOptions) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.execCalledWith = append(f.execCalledWith, opts)

	return f.execResp, f.execErr
}

func (f *fakeDocker) IsContainerRunning(_ context.Context, _ string) (bool, bool, error) {
	return false, false, nil
}

func (f *fakeDocker) Client() client.APIClient {
	return f.clientAPI
}

// withClient sets the value Client() returns; see the clientAPI field's
// comment. Defaults to nil (existing behavior for every other job's tests).
func (f *fakeDocker) withClient(api client.APIClient) *fakeDocker {
	f.clientAPI = api
	return f
}

func (f *fakeDocker) ContainerCreate(
	_ context.Context, _ *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *v1.Platform, _ string,
) (container.CreateResponse, error) {
	return f.containerCreateResp, f.containerCreateErr
}

func (f *fakeDocker) Stats(_ context.Context, _ string) (domain.ContainerStats, error) {
	return domain.ContainerStats{}, nil
}

// fakeNodeClients is a minimal node_clients.NodeClients wrapping a
// fakeDocker, for jobs (like createScratchContainerJob) that depend on the
// full NodeClients container but only ever call Docker() on it.
type fakeNodeClients struct {
	docker      *fakeDocker
	localState  node_clients.StateManager
	portManager node_clients.PortManager
}

func newFakeNodeClients(docker *fakeDocker) *fakeNodeClients {
	return &fakeNodeClients{docker: docker}
}

// withPortManager attaches a real ports.PortManager (in-memory, no external
// dependency beyond a local net.Listen availability probe) so jobs like
// upgrade_smerd.go's pauseOldContainerJob/prepareUpgradeVervConfigJob can be
// exercised at unit level without a hand-written PortManager fake.
func (f *fakeNodeClients) withPortManager(pm node_clients.PortManager) *fakeNodeClients {
	f.portManager = pm
	return f
}

func (f *fakeNodeClients) Docker() node_clients.Docker {
	return f.docker
}

func (f *fakeNodeClients) PortManager() node_clients.PortManager {
	return f.portManager
}

func (f *fakeNodeClients) PortManagerContainer() *ports.Container {
	return nil
}

func (f *fakeNodeClients) LocalStateManager() node_clients.StateManager {
	return f.localState
}

func (f *fakeNodeClients) HardwareManager() node_clients.HardwareManager {
	return nil
}

// fakeStateManager is a minimal in-memory implementation of
// node_clients.StateManager for exercising enable_statefull's jobs without
// touching disk - the real local_state.Manager persists to a JSON file on
// every Set/SetAndRelease call.
type fakeStateManager struct {
	mu    sync.Mutex
	state local_state.State
}

func newFakeStateManager(initial local_state.State) *fakeStateManager {
	return &fakeStateManager{state: initial}
}

func (f *fakeStateManager) Start() error { return nil }
func (f *fakeStateManager) Stop() error  { return nil }

func (f *fakeStateManager) Set(st local_state.State) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.state = st
}

func (f *fakeStateManager) Get() local_state.State {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.state
}

// GetForUpdate/SetAndRelease mirror local_state.Manager's lock/unlock
// contract: GetForUpdate acquires the lock and SetAndRelease releases it.
func (f *fakeStateManager) GetForUpdate() local_state.State {
	f.mu.Lock()

	return f.state
}

func (f *fakeStateManager) SetAndRelease(st local_state.State) {
	f.state = st
	f.mu.Unlock()
}

func (f *fakeStateManager) ValidateVelezPrivateKey(_ string) bool { return false }

// fakeClusterStorage is a minimal in-memory implementation of
// storage.Storage (== cluster_clients.ClusterStateManager) for exercising
// enable_statefull's update_cluster_state/init_node_storage jobs without a
// real Postgres-backed cluster state. Only Nodes() is configurable; the
// other accessors return nil since those jobs never call them.
type fakeClusterStorage struct {
	nodes storage.NodesStorage
}

func (f *fakeClusterStorage) Nodes() storage.NodesStorage                             { return f.nodes }
func (f *fakeClusterStorage) Services() storage.ServicesStorage                       { return nil }
func (f *fakeClusterStorage) Deployments() storage.DeploymentsStorage                 { return nil }
func (f *fakeClusterStorage) Plugins() storage.PluginsStorage                         { return nil }
func (f *fakeClusterStorage) ServiceDependencies() storage.ServiceDependenciesStorage { return nil }
func (f *fakeClusterStorage) ServiceResources() storage.ServiceResourcesStorage       { return nil }
func (f *fakeClusterStorage) Tasks() storage.TasksStorage                             { return nil }
func (f *fakeClusterStorage) Jobs() storage.JobsStorage                               { return nil }
func (f *fakeClusterStorage) TxManager() *sqldb.TxManager                             { return nil }

// fakeNodesStorage is a minimal in-memory implementation of
// storage.NodesStorage for exercising init_node_storage's InitNode call.
type fakeNodesStorage struct {
	mu            sync.Mutex
	initNodeErr   error
	initNodeCalls int
}

func (f *fakeNodesStorage) InitNode(_ context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.initNodeCalls++

	return f.initNodeErr
}

func (f *fakeNodesStorage) UpdateOnline(_ context.Context) error { return nil }

func (f *fakeNodesStorage) List(_ context.Context, _ domain.ListNodesReq) (domain.NodesList, error) {
	return domain.NodesList{}, nil
}

// fakeContainerAPI is a minimal hand-written fake for the narrow docker
// client.APIClient slices copy_to_volume's startLoaderContainerJob (startAPI)
// and copyFileJob (copyAPI) depend on - container start/stop and tar-based
// copy - so those jobs are unit-testable without implementing the full
// (100+ method) client.APIClient interface and without minimock.
type fakeContainerAPI struct {
	// client.APIClient is embedded as a nil interface value purely so
	// fakeContainerAPI satisfies the full (100+ method) interface and can be
	// injected via fakeDocker.withClient - only the methods explicitly
	// defined below are actually safe to call; anything else panics on the
	// nil embedded value, same as calling a method on a nil interface.
	client.APIClient

	mu sync.Mutex

	startErr        error
	startCalledWith []string

	stopErr        error
	stopCalledWith []string

	copyErr error
	// copyErrOnCall, if non-zero, makes CopyToContainer return copyErr only
	// on its Nth call (1-indexed) and succeed on every other call - used to
	// simulate one file failing mid-loop while earlier files already
	// succeeded.
	copyErrOnCall  int
	copyCalledWith []fakeCopyCall

	inspectResp container.InspectResponse
	inspectErr  error

	pauseErr        error
	pauseCalledWith []string

	unpauseErr        error
	unpauseCalledWith []string

	renameErr        error
	renameCalledWith []fakeRenameCall

	networkDisconnectErr        error
	networkDisconnectCalledWith []string

	networkConnectErr        error
	networkConnectCalledWith []string

	networkListResp []network.Summary
	networkListErr  error

	networkCreateErr        error
	networkCreateCalledWith []string

	copyFromResp []byte
	copyFromErr  error
}

type fakeRenameCall struct {
	containerID string
	newName     string
}

type fakeCopyCall struct {
	containerID string
	dstPath     string
	content     []byte
}

func newFakeContainerAPI() *fakeContainerAPI {
	return &fakeContainerAPI{}
}

func (f *fakeContainerAPI) ContainerStart(_ context.Context, containerID string, _ container.StartOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.startCalledWith = append(f.startCalledWith, containerID)

	return f.startErr
}

func (f *fakeContainerAPI) ContainerStop(_ context.Context, containerID string, _ container.StopOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.stopCalledWith = append(f.stopCalledWith, containerID)

	return f.stopErr
}

func (f *fakeContainerAPI) ContainerInspect(_ context.Context, _ string) (container.InspectResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.inspectResp, f.inspectErr
}

func (f *fakeContainerAPI) CopyToContainer(
	_ context.Context, containerID string, dstPath string, content io.Reader, _ container.CopyToContainerOptions,
) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	raw, err := io.ReadAll(content)
	if err != nil {
		return err
	}

	call := fakeCopyCall{
		containerID: containerID,
		dstPath:     dstPath,
		content:     raw,
	}
	f.copyCalledWith = append(f.copyCalledWith, call)

	if f.copyErrOnCall != 0 {
		if len(f.copyCalledWith) == f.copyErrOnCall {
			return f.copyErr
		}

		return nil
	}

	return f.copyErr
}

func (f *fakeContainerAPI) ContainerPause(_ context.Context, containerID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.pauseCalledWith = append(f.pauseCalledWith, containerID)

	return f.pauseErr
}

func (f *fakeContainerAPI) ContainerUnpause(_ context.Context, containerID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.unpauseCalledWith = append(f.unpauseCalledWith, containerID)

	return f.unpauseErr
}

func (f *fakeContainerAPI) ContainerRename(_ context.Context, containerID, newName string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.renameCalledWith = append(f.renameCalledWith, fakeRenameCall{containerID: containerID, newName: newName})

	return f.renameErr
}

func (f *fakeContainerAPI) NetworkDisconnect(_ context.Context, networkID, _ string, _ bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.networkDisconnectCalledWith = append(f.networkDisconnectCalledWith, networkID)

	return f.networkDisconnectErr
}

func (f *fakeContainerAPI) NetworkConnect(_ context.Context, networkID, _ string, _ *network.EndpointSettings) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.networkConnectCalledWith = append(f.networkConnectCalledWith, networkID)

	return f.networkConnectErr
}

func (f *fakeContainerAPI) NetworkList(_ context.Context, _ network.ListOptions) ([]network.Summary, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.networkListResp, f.networkListErr
}

func (f *fakeContainerAPI) NetworkCreate(_ context.Context, name string, _ network.CreateOptions) (network.CreateResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.networkCreateCalledWith = append(f.networkCreateCalledWith, name)

	return network.CreateResponse{}, f.networkCreateErr
}

// CopyFromContainer returns copyFromResp wrapped as a single-entry tar
// stream, mirroring what the real Docker engine returns for a file path -
// getConfigFromScratchContainerJob's readFileFromContainer unwraps that tar
// layer itself, same as dockerutils.ReadFromContainer does.
func (f *fakeContainerAPI) CopyFromContainer(
	_ context.Context, _ string, _ string,
) (io.ReadCloser, container.PathStat, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.copyFromErr != nil {
		return nil, container.PathStat{}, f.copyFromErr
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	_ = tw.WriteHeader(&tar.Header{Name: "content", Size: int64(len(f.copyFromResp))})
	_, _ = tw.Write(f.copyFromResp)
	_ = tw.Close()

	return io.NopCloser(&buf), container.PathStat{}, nil
}

// fakeVpnClient is a minimal in-memory implementation of
// cluster_clients.VervClosedNetworkClient for exercising
// connect_service_to_vpn's namespace/client-key jobs without a real
// Headscale server. Only the methods those jobs actually call are
// configurable; the rest return zero values since the jobs under test never
// invoke them.
type fakeVpnClient struct {
	mu sync.Mutex

	namespaces map[string]domain.VcnNamespace

	getNamespaceErr    error
	createNamespaceErr error

	authKeyResp domain.VcnAuthKey
	authKeyErr  error

	issuedKey   string
	issueKeyErr error
	issueCalls  int
}

func newFakeVpnClient() *fakeVpnClient {
	return &fakeVpnClient{namespaces: make(map[string]domain.VcnNamespace)}
}

func (f *fakeVpnClient) CreateNamespace(_ context.Context, name string) (domain.VcnNamespace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.createNamespaceErr != nil {
		return domain.VcnNamespace{}, f.createNamespaceErr
	}

	ns := domain.VcnNamespace{Id: "ns-" + name, Name: name}
	f.namespaces[name] = ns

	return ns, nil
}

func (f *fakeVpnClient) GetNamespace(_ context.Context, name string) (domain.VcnNamespace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.getNamespaceErr != nil {
		return domain.VcnNamespace{}, f.getNamespaceErr
	}

	return f.namespaces[name], nil
}

func (f *fakeVpnClient) ListNamespaces(_ context.Context) ([]domain.VcnNamespace, error) {
	return nil, nil
}

func (f *fakeVpnClient) DeleteNamespace(_ context.Context, _ string) error {
	return nil
}

func (f *fakeVpnClient) GetClientAuthKey(_ context.Context, _ domain.GetVcnAuthKeyReq) (domain.VcnAuthKey, error) {
	return f.authKeyResp, f.authKeyErr
}

func (f *fakeVpnClient) IssueClientKey(_ context.Context, _ domain.IssueClientKey) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.issueCalls++

	return f.issuedKey, f.issueKeyErr
}

func (f *fakeVpnClient) RegisterNode(_ context.Context, _ domain.RegisterVcnNodeReq) error {
	return nil
}

// fakeServiceDiscovery is a minimal in-memory implementation of
// cluster_clients.ServiceDiscovery (== makosh_be.MakoshBeAPIClient) for
// exercising connect_service_to_vpn's add_makosh_record job without a real
// makosh server.
type fakeServiceDiscovery struct {
	mu sync.Mutex

	upsertErr        error
	upsertCalledWith []*makosh_be.UpsertEndpoints_Request
}

func newFakeServiceDiscovery() *fakeServiceDiscovery {
	return &fakeServiceDiscovery{}
}

func (f *fakeServiceDiscovery) Version(_ context.Context, _ *makosh_be.Version_Request, _ ...grpc.CallOption) (*makosh_be.Version_Response, error) {
	return &makosh_be.Version_Response{}, nil
}

func (f *fakeServiceDiscovery) ListEndpoints(_ context.Context, _ *makosh_be.ListEndpoints_Request, _ ...grpc.CallOption) (*makosh_be.ListEndpoints_Response, error) {
	return &makosh_be.ListEndpoints_Response{}, nil
}

func (f *fakeServiceDiscovery) UpsertEndpoints(_ context.Context, in *makosh_be.UpsertEndpoints_Request, _ ...grpc.CallOption) (*makosh_be.UpsertEndpoints_Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.upsertCalledWith = append(f.upsertCalledWith, in)

	if f.upsertErr != nil {
		return nil, f.upsertErr
	}

	return &makosh_be.UpsertEndpoints_Response{}, nil
}

// fakeContainerService is a minimal in-memory implementation of
// service.ContainerService for exercising upgrade_smerd's
// checkSelfUpgradeJob/captureOldContainerJob without a real Docker-backed
// service layer. Only InspectSmerd is configurable; the rest return zero
// values since the jobs under test never invoke them.
type fakeContainerService struct {
	mu sync.Mutex

	inspectResp *velez_api.Smerd
	inspectErr  error

	inspectCalledWith []string
}

func newFakeContainerService() *fakeContainerService {
	return &fakeContainerService{}
}

func (f *fakeContainerService) ListSmerds(_ context.Context, _ *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error) {
	return &velez_api.ListSmerds_Response{}, nil
}

func (f *fakeContainerService) DropSmerds(_ context.Context, _ *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error) {
	return &velez_api.DropSmerd_Response{}, nil
}

func (f *fakeContainerService) InspectSmerd(_ context.Context, contId string) (*velez_api.Smerd, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.inspectCalledWith = append(f.inspectCalledWith, contId)

	return f.inspectResp, f.inspectErr
}

func (f *fakeContainerService) ConnectToNetwork(_ context.Context, _ domain.Connection) error {
	return nil
}

func (f *fakeContainerService) DisconnectFromNetwork(_ context.Context, _ domain.Connection) error {
	return nil
}

// fakeConfigurationService is a minimal in-memory implementation of
// service.ConfigurationService for exercising upgrade_smerd's
// fetchUpgradeConfigJob without a real Matreshka instance. Only
// GetEnvFromApi is configurable; the rest return zero values since the jobs
// under test never invoke them.
type fakeConfigurationService struct {
	mu sync.Mutex

	envResp *evon.Node
	envErr  error

	envCalledWith []domain.ConfigMeta
}

func newFakeConfigurationService() *fakeConfigurationService {
	return &fakeConfigurationService{}
}

func (f *fakeConfigurationService) GetVervFromApi(_ context.Context, _ domain.ConfigMeta) (matreshka.AppConfig, error) {
	return matreshka.AppConfig{}, nil
}

func (f *fakeConfigurationService) GetEnvFromApi(_ context.Context, meta domain.ConfigMeta) (*evon.Node, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.envCalledWith = append(f.envCalledWith, meta)

	if f.envResp == nil && f.envErr == nil {
		// Matches configurator.Configurator.GetEnvFromApi's real contract:
		// an empty (not-found) config still comes back as a non-nil zero
		// Node, never a bare nil - evon.NodeStorage.AddNode panics on nil.
		return &evon.Node{}, nil
	}

	return f.envResp, f.envErr
}

func (f *fakeConfigurationService) UpdateConfig(_ context.Context, _ domain.AppConfig) error {
	return nil
}

func (f *fakeConfigurationService) GetPlainFromApi(_ context.Context, _ domain.ConfigMeta) ([]byte, error) {
	return nil, nil
}

func (f *fakeConfigurationService) SubscribeOnChanges(_ ...string) error {
	return nil
}

func (f *fakeConfigurationService) UnsubscribeFromChanges(_ ...string) error {
	return nil
}

func (f *fakeConfigurationService) GetUpdates() <-chan domain.ConfigurationPatch {
	return nil
}
