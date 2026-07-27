package velez_api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testVervConfigName = "verv-cfg"
)

// These tests replace create_smerd_request_json_test.go's intent after the
// CreateSmerd_Request.config oneof (MatreshkaConfigSpec verv / PlainConfigSpec
// plain) was removed at the proto level in favor of two independent fields:
// optional MatreshkaConfigSpec verv and repeated FileConfig plain. Both are
// now plain protoc-gen-go struct fields (no Go interface field involved), so
// plain encoding/json round-trips them correctly with zero hand-written
// Marshal/Unmarshal code - these tests prove exactly that, the same
// round-trip internal/jobs relies on when checkpointing TaskContext
// (internal/jobs/checkpoint.go, internal/jobs/worker.go).

func TestCreateSmerdRequest_VervSurvivesRoundTrip(t *testing.T) {
	configName := testVervConfigName
	req := &CreateSmerd_Request{
		Name: "verv-test",
		Verv: &MatreshkaConfigSpec{ConfigName: &configName},
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

func TestCreateSmerdRequest_PlainSurvivesRoundTrip(t *testing.T) {
	req := &CreateSmerd_Request{
		Name: "loki",
		Plain: []*FileConfig{
			{Path: "/etc/loki/local-config.yaml", Content: []byte("auth_enabled: false")},
		},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var restored CreateSmerd_Request

	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	require.Len(t, restored.GetPlain(), 1)
	require.Equal(t, req.GetPlain()[0].GetPath(), restored.GetPlain()[0].GetPath())
	require.Equal(t, req.GetPlain()[0].GetContent(), restored.GetPlain()[0].GetContent())
	require.Equal(t, req.GetName(), restored.GetName())
}

// TestCreateSmerdRequest_VervAndPlainBothSurviveRoundTrip proves verv and
// plain are independent fields that can both be set and both survive a
// round-trip simultaneously - the behavior the oneof used to forbid.
func TestCreateSmerdRequest_VervAndPlainBothSurviveRoundTrip(t *testing.T) {
	configName := testVervConfigName
	req := &CreateSmerd_Request{
		Name: "combo",
		Verv: &MatreshkaConfigSpec{ConfigName: &configName},
		Plain: []*FileConfig{
			{Path: "/etc/app/extra.conf", Content: []byte("extra=1")},
		},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var restored CreateSmerd_Request

	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	require.NotNil(t, restored.GetVerv())
	require.Equal(t, configName, restored.GetVerv().GetConfigName())
	require.Len(t, restored.GetPlain(), 1)
	require.Equal(t, "/etc/app/extra.conf", restored.GetPlain()[0].GetPath())
}

func TestCreateSmerdRequest_NilVervAndPlainStayNil(t *testing.T) {
	req := &CreateSmerd_Request{Name: "no-config"}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var restored CreateSmerd_Request

	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	require.Nil(t, restored.GetVerv())
	require.Empty(t, restored.GetPlain())
	require.Equal(t, req.GetName(), restored.GetName())
}

// TestCreateSmerdRequest_SurvivesRoundTripEmbeddedInTaskPayload mirrors how
// internal/jobs actually persists this type: nested inside a *TaskPayload,
// marshaled/unmarshaled as a whole via encoding/json (see
// internal/jobs/engine.go's Enqueue/Watch and worker.go's run()).
func TestCreateSmerdRequest_SurvivesRoundTripEmbeddedInTaskPayload(t *testing.T) {
	configName := testVervConfigName
	payload := &CreateSmerdTaskPayload{
		Request: &CreateSmerd_Request{
			Name:      "loki",
			ImageName: "grafana/loki:main",
			Verv:      &MatreshkaConfigSpec{ConfigName: &configName},
			Plain: []*FileConfig{
				{Path: "a.yaml", Content: []byte("x")},
			},
		},
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored CreateSmerdTaskPayload

	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	require.NotNil(t, restored.GetRequest().GetVerv())
	require.Equal(t, configName, restored.GetRequest().GetVerv().GetConfigName())
	require.Len(t, restored.GetRequest().GetPlain(), 1)
	require.Equal(t, payload.GetRequest().GetName(), restored.GetRequest().GetName())
	require.Equal(t, payload.GetRequest().GetImageName(), restored.GetRequest().GetImageName())
}

// TestUpgradeSmerdRequest_SurvivesRoundTripEmbeddedInTaskPayload covers the
// second embedder of CreateSmerd_Request: UpgradeSmerdTaskPayload.Request
// (internal/jobs/upgrade_smerd.go), which shares the same at-risk shape.
func TestUpgradeSmerdRequest_SurvivesRoundTripEmbeddedInTaskPayload(t *testing.T) {
	configName := testVervConfigName
	payload := &UpgradeSmerdTaskPayload{
		Request: &CreateSmerd_Request{
			Name: "mysvc",
			Verv: &MatreshkaConfigSpec{ConfigName: &configName},
		},
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored UpgradeSmerdTaskPayload

	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	require.NotNil(t, restored.GetRequest().GetVerv())
	require.Equal(t, configName, restored.GetRequest().GetVerv().GetConfigName())
	require.Equal(t, payload.GetRequest().GetName(), restored.GetRequest().GetName())
}
