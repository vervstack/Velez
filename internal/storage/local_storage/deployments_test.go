package local_storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/sqlc-dev/pqtype"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/tests/test_helper"
)

// TestDeployments_ExecuteSerializesAgainstList proves Execute and List share
// the same lock (d.mu), not two independent ones: List can't complete while
// Execute's fn is still running. Without this, a caller of executeDeployment
// (verv_services.VervService) running the composite spec+deployment write
// through Execute would get no real exclusion against a concurrent
// deploy_watcher tick calling List().
func TestDeployments_ExecuteSerializesAgainstList(t *testing.T) {
	d := newDeploymentsStorage(test_helper.NewRealDocker(t))

	started := make(chan struct{})
	listDone := make(chan struct{})

	go func() {
		<-started

		_, err := d.List(context.Background(), domain.ListDeploymentsReq{})
		require.NoError(t, err)

		close(listDone)
	}()

	var listFinishedDuringExecute bool

	err := d.Execute(func(_ *sql.Tx) error {
		close(started)

		select {
		case <-listDone:
			listFinishedDuringExecute = true
		case <-time.After(100 * time.Millisecond):
			// Expected: List() is blocked on d.mu until Execute returns.
		}

		return nil
	})
	require.NoError(t, err)

	require.False(t, listFinishedDuringExecute, "List() must not complete while Execute holds the lock")

	<-listDone
}

// TestDeployments_WithTxDoesNotDeadlock proves the Querier returned by
// WithTx (used from inside Execute's fn) can call CreateSpecification/
// CreateDeployment/GetSpecificationById/UpdateDeploymentStatus without
// re-locking d.mu - the deadlock hazard flagged for this design, since
// sync.Mutex isn't reentrant and Execute already holds the lock when it
// calls fn.
func TestDeployments_WithTxDoesNotDeadlock(t *testing.T) {
	d := newDeploymentsStorage(test_helper.NewRealDocker(t))
	ctx := context.Background()

	done := make(chan error, 1)

	go func() {
		done <- d.Execute(func(tx *sql.Tx) error {
			q := d.WithTx(tx)

			specId, err := q.CreateSpecification(ctx, deployments_queries.CreateSpecificationParams{Name: "spec"})
			if err != nil {
				return rerrors.Wrap(err, "error creating specification")
			}

			_, err = q.CreateDeployment(ctx, deployments_queries.CreateDeploymentParams{
				NodeID: 1,
				Status: deployments_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT,
				SpecID: specId,
			})
			if err != nil {
				return rerrors.Wrap(err, "error creating deployment")
			}

			_, err = q.GetSpecificationById(ctx, specId)
			if err != nil {
				return rerrors.Wrap(err, "error getting specification")
			}

			return q.UpdateDeploymentStatus(ctx, deployments_queries.UpdateDeploymentStatusParams{
				ID:     1,
				Status: deployments_queries.VelezDeploymentStatusRUNNING,
			})
		})
	}()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Execute deadlocked calling into its own WithTx querier")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	require.Len(t, d.deployments, 1)
	require.Equal(t, deployments_queries.VelezDeploymentStatusRUNNING, d.deployments[0].Status)
}

// TestDeployments_ExecuteCompositeWriteIsAtomic stresses the composite
// spec+deployment write (as verv_services.executeDeployment performs it)
// against concurrent readers calling List(). Every write's spec and
// deployment go in under one Execute call, so a reader must never observe a
// spec count ahead of the deployment count it should always match one-to-
// one - this is the atomicity gap described for the nil-TxManager special
// case this design replaces. Run with -race to also confirm there's no data
// race between the writer and the readers.
func TestDeployments_ExecuteCompositeWriteIsAtomic(t *testing.T) {
	d := newDeploymentsStorage(test_helper.NewRealDocker(t))
	ctx := context.Background()

	const writes = 300

	stop := make(chan struct{})
	violation := make(chan string, 1)

	var readers sync.WaitGroup

	for range 4 {
		readers.Add(1)

		go func() {
			defer readers.Done()

			for {
				select {
				case <-stop:
					return
				default:
				}

				d.mu.Lock()

				specs := len(d.specs)
				deps := len(d.deployments)
				d.mu.Unlock()

				if specs != deps {
					select {
					case violation <- "observed specs/deployments counts diverge mid-write":
					default:
					}

					return
				}

				_, err := d.List(ctx, domain.ListDeploymentsReq{})
				require.NoError(t, err)
			}
		}()
	}

	for range writes {
		err := d.Execute(func(tx *sql.Tx) error {
			q := d.WithTx(tx)

			specId, specErr := q.CreateSpecification(ctx, deployments_queries.CreateSpecificationParams{Name: "spec"})
			if specErr != nil {
				return rerrors.Wrap(specErr, "error creating specification")
			}

			_, specErr = q.CreateDeployment(ctx, deployments_queries.CreateDeploymentParams{
				NodeID: 1,
				Status: deployments_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT,
				SpecID: specId,
			})
			if specErr != nil {
				return rerrors.Wrap(specErr, "error creating deployment")
			}

			return nil
		})
		require.NoError(t, err)
	}

	close(stop)
	readers.Wait()

	select {
	case msg := <-violation:
		t.Fatal(msg)
	default:
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	require.Len(t, d.deployments, writes)
}

// TestDeployments_GetSpecificationById_PrefersLiveContainerOverStaleSpec
// proves that once a service's container exists, GetSpecificationById
// derives the spec from it instead of from the recorded map entry, which
// still holds what was originally requested (a different, "stale" image
// here) - the container is what upgrade correctness must diff against.
func TestDeployments_GetSpecificationById_PrefersLiveContainerOverStaleSpec(t *testing.T) {
	t.Parallel()

	cli := test_helper.NewRealDockerAPI(t)
	test_helper.EnsurePulled(t, cli, test_helper.HelloWorldAppImage)

	name := test_helper.UniqueName(t, "spec-from-container")

	cfg := &container.Config{
		Image: test_helper.HelloWorldAppImage,
		Env:   []string{"FOO=bar"},
		Labels: map[string]string{
			labels.VervServiceLabel: name,
		},
		Healthcheck: &container.HealthConfig{
			Test:     []string{"CMD-SHELL", "echo ok"},
			Interval: 5 * time.Second,
			Timeout:  2 * time.Second,
			Retries:  3,
		},
	}

	hostCfg := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name:              container.RestartPolicyOnFailure,
			MaximumRetryCount: 4,
		},
		PortBindings: nat.PortMap{
			"8080/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "18099"}},
		},
	}

	created, err := cli.ContainerCreate(context.Background(), cfg, hostCfg, nil, nil, name)
	require.NoError(t, err)

	t.Cleanup(func() {
		test_helper.RemoveContainer(t, cli, created.ID)
	})

	d := newDeploymentsStorage(test_helper.NewRealDocker(t))
	ctx := context.Background()

	staleReq := &pb.CreateSmerd_Request{Name: name, ImageName: "stale-image:v0"}

	stalePayload, err := json.Marshal(staleReq)
	require.NoError(t, err)

	specParams := deployments_queries.CreateSpecificationParams{
		Name:        "spec-name",
		VervPayload: pqtype.NullRawMessage{RawMessage: stalePayload, Valid: true},
	}

	specId, err := d.CreateSpecification(ctx, specParams)
	require.NoError(t, err)

	row, err := d.GetSpecificationById(ctx, specId)
	require.NoError(t, err)

	smerdReq := &pb.CreateSmerd_Request{}

	err = json.Unmarshal(row.VervPayload.RawMessage, smerdReq)
	require.NoError(t, err)

	require.Equal(t, test_helper.HelloWorldAppImage, smerdReq.GetImageName())
	require.Equal(t, "bar", smerdReq.GetEnv()["FOO"])
	require.Equal(t, "echo ok", smerdReq.GetHealthcheck().GetCommand())
	require.Equal(t, uint32(5), smerdReq.GetHealthcheck().GetIntervalSecond())
	require.Equal(t, pb.RestartPolicyType_on_failure, smerdReq.GetRestart().GetType())
	require.Equal(t, uint32(4), smerdReq.GetRestart().GetFailureCount())
	require.Len(t, smerdReq.GetSettings().GetPorts(), 1)
	require.Equal(t, uint32(18099), smerdReq.GetSettings().GetPorts()[0].GetExposedTo())
}

// TestDeployments_GetSpecificationById_FallsBackToMapBeforeContainerExists
// proves the pre-container bridge window still works: with no live
// container backing name yet, GetSpecificationById returns the recorded map
// entry as-is, mirroring dockerPgInstances/dockerSecrets's pending overlays.
func TestDeployments_GetSpecificationById_FallsBackToMapBeforeContainerExists(t *testing.T) {
	t.Parallel()

	d := newDeploymentsStorage(test_helper.NewRealDocker(t))
	ctx := context.Background()

	name := test_helper.UniqueName(t, "pending-spec")

	req := &pb.CreateSmerd_Request{Name: name, ImageName: "pending-image:v0"}

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	specParams := deployments_queries.CreateSpecificationParams{
		Name:        "spec-name",
		VervPayload: pqtype.NullRawMessage{RawMessage: payload, Valid: true},
	}

	specId, err := d.CreateSpecification(ctx, specParams)
	require.NoError(t, err)

	row, err := d.GetSpecificationById(ctx, specId)
	require.NoError(t, err)

	smerdReq := &pb.CreateSmerd_Request{}

	err = json.Unmarshal(row.VervPayload.RawMessage, smerdReq)
	require.NoError(t, err)

	require.Equal(t, "pending-image:v0", smerdReq.GetImageName())
}

func startLabelledContainer(t *testing.T, name string) {
	t.Helper()

	cli := test_helper.NewRealDockerAPI(t)
	test_helper.EnsurePulled(t, cli, test_helper.HelloWorldAppImage)

	cfg := &container.Config{
		Image:  test_helper.HelloWorldAppImage,
		Labels: map[string]string{labels.VervServiceLabel: name},
	}

	created, err := cli.ContainerCreate(context.Background(), cfg, nil, nil, nil, name)
	require.NoError(t, err)

	t.Cleanup(func() {
		test_helper.RemoveContainer(t, cli, created.ID)
	})

	err = cli.ContainerStart(context.Background(), created.ID, container.StartOptions{})
	require.NoError(t, err)
}

func newNamedSpecParams(t *testing.T, name string) deployments_queries.CreateSpecificationParams {
	t.Helper()

	req := &pb.CreateSmerd_Request{Name: name, ImageName: test_helper.HelloWorldAppImage}

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	return deployments_queries.CreateSpecificationParams{
		Name:        name,
		VervPayload: pqtype.NullRawMessage{RawMessage: payload, Valid: true},
	}
}

func createMemoryDeployment(
	ctx context.Context,
	t *testing.T,
	d *deployments,
	name string,
	status deployments_queries.VelezDeploymentStatus,
) {
	t.Helper()

	specId, err := d.CreateSpecification(ctx, newNamedSpecParams(t, name))
	require.NoError(t, err)

	depParams := deployments_queries.CreateDeploymentParams{
		NodeID: domain.SelfNodeId,
		Status: status,
		SpecID: specId,
	}

	_, err = d.CreateDeployment(ctx, depParams)
	require.NoError(t, err)
}

func Test_Deployments_DerivesRunningRowFromContainerAfterRestart(t *testing.T) {
	t.Parallel()

	name := test_helper.UniqueName(t, "derived-dep")
	startLabelledContainer(t, name)

	d := newDeploymentsStorage(test_helper.NewRealDocker(t))
	ctx := context.Background()

	req := domain.ListDeploymentsReq{ServiceName: name}

	rows, err := d.List(ctx, req)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, deployments_queries.VelezDeploymentStatusRUNNING, rows[0].Status)
	require.Negative(t, rows[0].Id)
	require.Equal(t, rows[0].Id, rows[0].SpecId)

	spec, err := d.GetSpecificationById(ctx, rows[0].SpecId)
	require.NoError(t, err)
	require.Equal(t, name, spec.Name)
	require.Equal(t, rows[0].SpecId, spec.ID)

	smerdReq := &pb.CreateSmerd_Request{}

	err = json.Unmarshal(spec.VervPayload.RawMessage, smerdReq)
	require.NoError(t, err)
	require.Equal(t, name, smerdReq.GetName())
}

func Test_Deployments_MemoryRunningRowSuppressesDerivedRow(t *testing.T) {
	t.Parallel()

	name := test_helper.UniqueName(t, "memory-wins")
	startLabelledContainer(t, name)

	d := newDeploymentsStorage(test_helper.NewRealDocker(t))
	ctx := context.Background()

	createMemoryDeployment(ctx, t, d, name, deployments_queries.VelezDeploymentStatusRUNNING)

	rows, err := d.List(ctx, domain.ListDeploymentsReq{ServiceName: name})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Positive(t, rows[0].Id)
}

func Test_Deployments_DerivedRowReappearsWhenMemoryRowsFailed(t *testing.T) {
	t.Parallel()

	name := test_helper.UniqueName(t, "self-heal")
	startLabelledContainer(t, name)

	d := newDeploymentsStorage(test_helper.NewRealDocker(t))
	ctx := context.Background()

	createMemoryDeployment(ctx, t, d, name, deployments_queries.VelezDeploymentStatusFAILED)

	rows, err := d.List(ctx, domain.ListDeploymentsReq{ServiceName: name})
	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Equal(t, deployments_queries.VelezDeploymentStatusFAILED, rows[0].Status)
	require.Equal(t, deployments_queries.VelezDeploymentStatusRUNNING, rows[1].Status)
	require.Negative(t, rows[1].Id)
}

func Test_Deployments_UpdateStatusOnDerivedIdIsNoop(t *testing.T) {
	t.Parallel()

	d := newDeploymentsStorage(test_helper.NewRealDocker(t))

	arg := deployments_queries.UpdateDeploymentStatusParams{
		ID:     derivedId("some-service"),
		Status: deployments_queries.VelezDeploymentStatusFAILED,
	}

	err := d.UpdateDeploymentStatus(context.Background(), arg)
	require.NoError(t, err)
}
