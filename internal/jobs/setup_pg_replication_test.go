package jobs

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupPgReplicationSlotName(t *testing.T) {
	t.Parallel()

	require.Equal(t, "velez_pgaas_orders", pgReplicaSlotName("pgaas_orders"))
	require.NotEqual(t, pgReplicaSlotName("pgaas_orders"), pgReplicaSlotName("pgaas-orders"))
	require.NotEqual(t, pgReplicaSlotName("pgaas-orders"), pgReplicaSlotName("pgaas.Orders"))

	longName := "pgaas_" + strings.Repeat("order", 20)
	longSlot := pgReplicaSlotName(longName)

	require.Len(t, longSlot, pgReplicaSlotMaxLength)
	require.Regexp(t, "^velez_[a-z0-9_]+$", longSlot)
	require.Equal(t, longSlot, pgReplicaSlotName(longName))
	require.Equal(t, longSlot, pgReplicaApplicationName(longName))
	require.Equal(t, "pgaas_orders", pgReplicaApplicationName("pgaas_orders"))
	require.NotEqual(t, longSlot, pgReplicaSlotName(longName+"2"))
}

func TestSetupPgReplicationTargetEmptyQuery(t *testing.T) {
	t.Parallel()

	// PostgreSQL catalogs check all schemas in the connected database, including partitioned tables.
	require.Contains(t, pgTargetEmptyQuery, "pg_is_in_recovery()")
	require.Contains(t, pgTargetEmptyQuery, "c.relkind IN ('r', 'p', 'f', 'm')")
	require.Contains(t, pgTargetEmptyQuery, "n.nspname NOT IN ('pg_catalog', 'information_schema')")
	require.Contains(t, pgTargetEmptyQuery, "n.nspname !~ '^pg_(toast|temp_)'")
}

func TestPgReplicationStatusMapping(t *testing.T) {
	t.Parallel()

	lag := sql.NullInt64{Int64: 128, Valid: true}
	replay := sql.NullFloat64{Float64: 1.5, Valid: true}
	behind := pgReplicaStatus("pgaas_standby", "streaming", lag, replay, false)

	require.Equal(t, "pgaas_standby", behind.ApplicationName)
	require.Equal(t, "streaming", behind.State)
	require.EqualValues(t, 128, behind.LagBytes)
	require.EqualValues(t, 1.5, behind.ReplayLagSeconds)
	require.False(t, behind.InSync)

	noLag := sql.NullInt64{}
	noReplayLag := sql.NullFloat64{}
	sync := pgReplicaStatus("pgaas_standby", "streaming", noLag, noReplayLag, true)

	require.Zero(t, sync.LagBytes)
	require.Zero(t, sync.ReplayLagSeconds)
	require.True(t, sync.InSync)
}
