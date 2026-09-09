package local_storage

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
)

// TestDeployments_ExecuteSerializesAgainstList proves Execute and List share
// the same lock (d.mu), not two independent ones: List can't complete while
// Execute's fn is still running. Without this, a caller of executeDeployment
// (verv_services.VervService) running the composite spec+deployment write
// through Execute would get no real exclusion against a concurrent
// deploy_watcher tick calling List().
func TestDeployments_ExecuteSerializesAgainstList(t *testing.T) {
	d := newDeploymentsStorage()

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
	d := newDeploymentsStorage()
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

	list, err := d.ListDeployments(ctx, domain.ListDeploymentsReq{})
	require.NoError(t, err)
	require.Equal(t, uint64(1), list.Total)
	require.Equal(t, deployments_queries.VelezDeploymentStatusRUNNING, list.Deployments[0].Status)
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
	d := newDeploymentsStorage()
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

	list, err := d.ListDeployments(ctx, domain.ListDeploymentsReq{})
	require.NoError(t, err)
	require.Equal(t, uint64(writes), list.Total)
}
