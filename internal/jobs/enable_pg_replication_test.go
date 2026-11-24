package jobs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func Test_EnablePgReplication_RequiredChanges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings pgReplicationSettings
		want     []string
	}{
		{
			name:     "ready at minimum values",
			settings: pgReplicationSettings{"replica", "on", 10, 10},
		},
		{
			name:     "keep logical and higher limits",
			settings: pgReplicationSettings{pgWalLevelLogical, "on", 20, 15},
		},
		{
			name:     "change only inadequate settings",
			settings: pgReplicationSettings{pgWalLevelLogical, "off", 5, 20},
			want: []string{
				"ALTER SYSTEM SET wal_log_hints = on",
				"ALTER SYSTEM SET max_wal_senders = 10",
			},
		},
		{
			name:     "change all inadequate settings",
			settings: pgReplicationSettings{"minimal", "off", 0, 0},
			want: []string{
				"ALTER SYSTEM SET wal_level = 'replica'",
				"ALTER SYSTEM SET wal_log_hints = on",
				"ALTER SYSTEM SET max_wal_senders = 10",
				"ALTER SYSTEM SET max_replication_slots = 10",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.want, test.settings.changes())
		})
	}
}

func Test_EnablePgReplication_AlreadyAppliedSkipsWrites(t *testing.T) {
	t.Parallel()

	payload := &velez_api.EnablePgReplicationTaskPayload{
		InstanceName:  "pgaas_example",
		Applied:       true,
		RestartNeeded: true,
	}

	encoded, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored velez_api.EnablePgReplicationTaskPayload

	err = json.Unmarshal(encoded, &restored)
	require.NoError(t, err)
	require.Equal(t, payload, &restored)

	job := &pgReplicationReadiness{payload: &restored}

	err = job.applySettings(t.Context())
	require.NoError(t, err)
	require.True(t, restored.GetRestartNeeded())
}

func Test_EnablePgReplication_ReadyInstanceSkipsRestart(t *testing.T) {
	t.Parallel()

	payload := &velez_api.EnablePgReplicationTaskPayload{InstanceName: "pgaas_example", Applied: true}
	job := &pgReplicationReadiness{payload: payload}

	err := job.applySettings(t.Context())
	require.NoError(t, err)

	err = job.restartPg(t.Context())
	require.NoError(t, err)

	err = job.waitReady(t.Context())
	require.NoError(t, err)
}
