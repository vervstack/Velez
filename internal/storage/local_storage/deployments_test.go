package local_storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/sqlc-dev/pqtype"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
)

const (
	testSpecName = "spec-1"
)

func Test_deployments_CreateSpecification_GetSpecificationByIdRoundTrip(t *testing.T) {
	t.Parallel()

	d := newDeploymentsStorage()

	arg := deployments_queries.CreateSpecificationParams{
		Name:        testSpecName,
		VervPayload: pqtype.NullRawMessage{RawMessage: []byte(`{"a":1}`), Valid: true},
	}

	id, err := d.CreateSpecification(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, int64(1), id)

	row, err := d.GetSpecificationById(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, arg.Name, row.Name)
	require.Equal(t, arg.VervPayload, row.VervPayload)
}

func Test_deployments_GetSpecificationById_UnknownIdNotFound(t *testing.T) {
	t.Parallel()

	d := newDeploymentsStorage()

	_, err := d.GetSpecificationById(context.Background(), 999)
	require.ErrorIs(t, err, storage.ErrNotFound)
}

func Test_deployments_CreateDeployment_InheritsServiceIdFromSpec(t *testing.T) {
	t.Parallel()

	d := newDeploymentsStorage()

	specId, err := d.CreateSpecification(context.Background(), deployments_queries.CreateSpecificationParams{
		Name:      testSpecName,
		ServiceID: sql.NullInt64{Int64: 42, Valid: true},
	})
	require.NoError(t, err)

	_, err = d.CreateDeployment(context.Background(), deployments_queries.CreateDeploymentParams{
		NodeID: 1,
		Status: deployments_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT,
		SpecID: specId,
	})
	require.NoError(t, err)

	deployments, err := d.List(context.Background(), domain.ListDeploymentsReq{})
	require.NoError(t, err)
	require.Len(t, deployments, 1)
	require.Equal(t, int64(42), deployments[0].ServiceId)
	require.Equal(t, specId, deployments[0].SpecId)
}

func Test_deployments_List_FiltersByNotStatusAndNodeIds(t *testing.T) {
	t.Parallel()

	d := newDeploymentsStorage()

	seedDeployment(t, d, 1, deployments_queries.VelezDeploymentStatusRUNNING)
	seedDeployment(t, d, 1, deployments_queries.VelezDeploymentStatusFAILED)
	seedDeployment(t, d, 2, deployments_queries.VelezDeploymentStatusRUNNING)

	req := domain.ListDeploymentsReq{
		NodeIds:   []int64{1},
		NotStatus: []deployments_queries.VelezDeploymentStatus{deployments_queries.VelezDeploymentStatusFAILED},
	}

	deployments, err := d.List(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, deployments, 1)
	require.Equal(t, int64(1), deployments[0].NodeId)
	require.Equal(t, deployments_queries.VelezDeploymentStatusRUNNING, deployments[0].Status)
}

func Test_deployments_ListDeployments_ReportsTotalBeforePaging(t *testing.T) {
	t.Parallel()

	d := newDeploymentsStorage()

	seedDeployment(t, d, 1, deployments_queries.VelezDeploymentStatusRUNNING)
	seedDeployment(t, d, 1, deployments_queries.VelezDeploymentStatusRUNNING)
	seedDeployment(t, d, 1, deployments_queries.VelezDeploymentStatusRUNNING)

	req := domain.ListDeploymentsReq{
		Paging: domain.Paging{Limit: 2},
	}

	list, err := d.ListDeployments(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, list.Deployments, 2)
	require.Equal(t, uint64(3), list.Total)
}

func Test_deployments_UpdateDeploymentStatus(t *testing.T) {
	t.Parallel()

	d := newDeploymentsStorage()

	specParams := deployments_queries.CreateSpecificationParams{Name: testSpecName}

	specId, err := d.CreateSpecification(context.Background(), specParams)
	require.NoError(t, err)

	_, err = d.CreateDeployment(context.Background(), deployments_queries.CreateDeploymentParams{
		NodeID: 1,
		Status: deployments_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT,
		SpecID: specId,
	})
	require.NoError(t, err)

	err = d.UpdateDeploymentStatus(context.Background(), deployments_queries.UpdateDeploymentStatusParams{
		ID:     1,
		Status: deployments_queries.VelezDeploymentStatusRUNNING,
	})
	require.NoError(t, err)

	deployments, err := d.List(context.Background(), domain.ListDeploymentsReq{})
	require.NoError(t, err)
	require.Equal(t, deployments_queries.VelezDeploymentStatusRUNNING, deployments[0].Status)
}

func Test_deployments_UpdateDeploymentStatus_UnknownIdNotFound(t *testing.T) {
	t.Parallel()

	d := newDeploymentsStorage()

	err := d.UpdateDeploymentStatus(context.Background(), deployments_queries.UpdateDeploymentStatusParams{ID: 999})
	require.ErrorIs(t, err, storage.ErrNotFound)
}

func seedDeployment(
	t *testing.T,
	d *deployments,
	nodeId int32,
	status deployments_queries.VelezDeploymentStatus,
) {
	t.Helper()

	specParams := deployments_queries.CreateSpecificationParams{Name: testSpecName}

	specId, err := d.CreateSpecification(context.Background(), specParams)
	require.NoError(t, err)

	_, err = d.CreateDeployment(context.Background(), deployments_queries.CreateDeploymentParams{
		NodeID: nodeId,
		Status: status,
		SpecID: specId,
	})
	require.NoError(t, err)
}
