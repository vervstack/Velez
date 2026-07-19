package jobs

import (
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
	"go.vervstack.ru/makosh/pkg/makosh_be"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/ports"
	"go.vervstack.ru/Velez/internal/domain"
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
	return nil
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
	docker *fakeDocker
}

func newFakeNodeClients(docker *fakeDocker) *fakeNodeClients {
	return &fakeNodeClients{docker: docker}
}

func (f *fakeNodeClients) Docker() node_clients.Docker {
	return f.docker
}

func (f *fakeNodeClients) PortManager() node_clients.PortManager {
	return nil
}

func (f *fakeNodeClients) PortManagerContainer() *ports.Container {
	return nil
}

func (f *fakeNodeClients) LocalStateManager() node_clients.StateManager {
	return nil
}

func (f *fakeNodeClients) HardwareManager() node_clients.HardwareManager {
	return nil
}

// fakeContainerAPI is a minimal hand-written fake for the narrow docker
// client.APIClient slices copy_to_volume's startLoaderContainerJob (startAPI)
// and copyFileJob (copyAPI) depend on - container start/stop and tar-based
// copy - so those jobs are unit-testable without implementing the full
// (100+ method) client.APIClient interface and without minimock.
type fakeContainerAPI struct {
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
