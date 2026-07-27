package jobs

import (
	"context"
	"testing"

	"go.redsock.ru/evon"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

const (
	testFetchedEnvValue = "from-verv"
)

// TestFetchSmerdConfigJob_VervAndPlain_BothApply proves verv and plain are
// independent (no longer a mutually-exclusive oneof): setting both on the
// same request must apply both effects - verv's env fetch AND plain's file
// mounts - not just one of them.
func TestFetchSmerdConfigJob_VervAndPlain_BothApply(t *testing.T) {
	configService := newFakeConfigurationService()

	configService.envResp = &evon.Node{Name: "BAR", Value: testFetchedEnvValue}

	payload := &velez_api.CreateSmerdTaskPayload{
		Request: &velez_api.CreateSmerd_Request{
			Name: "combo-svc",
			Env:  map[string]string{},
			Verv: &velez_api.MatreshkaConfigSpec{},
			Plain: []*velez_api.FileConfig{
				{Path: "/etc/app/extra.conf", Content: []byte("extra=1")},
			},
		},
	}

	j := &fetchSmerdConfigJob{configService: configService, req: payload, imageMeta: payload, mounts: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetRequest().GetEnv()["BAR"] != testFetchedEnvValue {
		t.Errorf("expected verv's fetched env var to be applied, got %v", payload.GetRequest().GetEnv())
	}

	if string(payload.GetPathToFiles()["/etc/app/extra.conf"]) != "extra=1" {
		t.Errorf("expected plain's file to be mounted, got %v", payload.GetPathToFiles())
	}
}

// TestFetchSmerdConfigJob_SetEnv_CallerValueNotOverwritten proves setEnv's
// new precedence: a caller-supplied env value (already present in
// request.Env before Do runs, i.e. merged in earlier in the flow) must win
// over a colliding value verv's fetch would otherwise produce.
func TestFetchSmerdConfigJob_SetEnv_CallerValueNotOverwritten(t *testing.T) {
	configService := newFakeConfigurationService()

	configService.envResp = &evon.Node{Name: testEnvKeyFoo, Value: testFetchedEnvValue}

	payload := &velez_api.CreateSmerdTaskPayload{
		Request: &velez_api.CreateSmerd_Request{
			Name: "svc",
			Env:  map[string]string{testEnvKeyFoo: "from-caller"},
			Verv: &velez_api.MatreshkaConfigSpec{},
		},
	}

	j := &fetchSmerdConfigJob{configService: configService, req: payload, imageMeta: payload, mounts: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetRequest().GetEnv()[testEnvKeyFoo] != "from-caller" {
		t.Errorf("expected caller-supplied env value to win, got %v", payload.GetRequest().GetEnv())
	}
}
