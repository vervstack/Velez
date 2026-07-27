package velez_api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// These tests guard the hand-written MarshalJSON/UnmarshalJSON pair in
// create_smerd_request_json.go. Without them, CreateSmerd_Request.Config (a
// proto oneof, i.e. a Go interface field) marshals fine but can never
// unmarshal back into a concrete type - encoding/json has no way to
// allocate one for an interface field. This silently broke internal/jobs'
// create_smerd task engine (see docs/jobs_migration.md's oneof check and
// docs/plans/testing.md's 2026-07-19 progress log) because TaskContext is
// persisted via plain encoding/json, not protojson. Any future change to
// these methods, or to CreateSmerd_Request.Config's shape, must keep these
// tests green.

func TestCreateSmerdRequest_PlainConfigSurvivesRoundTrip(t *testing.T) {
	req := &CreateSmerd_Request{
		Name: "loki",
		Config: &CreateSmerd_Request_Plain{
			Plain: &PlainConfigSpec{
				Configs: map[string][]byte{"/etc/loki/local-config.yaml": []byte("auth_enabled: false")},
			},
		},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var restored CreateSmerd_Request

	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	require.NotNil(t, restored.GetPlain())
	require.Equal(t, req.GetPlain().GetConfigs(), restored.GetPlain().GetConfigs())
	require.Equal(t, req.GetName(), restored.GetName())
}

func TestCreateSmerdRequest_VervConfigSurvivesRoundTrip(t *testing.T) {
	configName := "verv-cfg"
	req := &CreateSmerd_Request{
		Name:   "verv-test",
		Config: &CreateSmerd_Request_Verv{Verv: &MatreshkaConfigSpec{ConfigName: &configName}},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var restored CreateSmerd_Request

	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	require.NotNil(t, restored.GetVerv())
	require.Equal(t, configName, restored.GetVerv().GetConfigName())
	require.Equal(t, req.GetName(), restored.GetName())
}

func TestCreateSmerdRequest_NilConfigStaysNil(t *testing.T) {
	req := &CreateSmerd_Request{Name: "no-config"}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var restored CreateSmerd_Request

	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	require.Nil(t, restored.GetConfig())
	require.Equal(t, req.GetName(), restored.GetName())
}

// TestCreateSmerdRequest_SurvivesRoundTripEmbeddedInTaskPayload mirrors how
// internal/jobs actually persists this type: nested inside a *TaskPayload,
// marshaled/unmarshaled as a whole via encoding/json (see
// internal/jobs/engine.go's Enqueue/Watch and worker.go's run()).
func TestCreateSmerdRequest_SurvivesRoundTripEmbeddedInTaskPayload(t *testing.T) {
	payload := &CreateSmerdTaskPayload{
		Request: &CreateSmerd_Request{
			Name:      "loki",
			ImageName: "grafana/loki:main",
			Config: &CreateSmerd_Request_Plain{
				Plain: &PlainConfigSpec{Configs: map[string][]byte{"a.yaml": []byte("x")}},
			},
		},
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored CreateSmerdTaskPayload

	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	require.NotNil(t, restored.GetRequest().GetPlain())
	require.Equal(t, payload.GetRequest().GetName(), restored.GetRequest().GetName())
	require.Equal(t, payload.GetRequest().GetImageName(), restored.GetRequest().GetImageName())
}
