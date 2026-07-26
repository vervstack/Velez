package container_service_task

import (
	"context"
	"sync"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// fakeAPIClient embeds client.APIClient as a nil interface value purely so
// this type satisfies the ~100-method interface without minimock; only
// ContainerInspect is overridden, since that's all IsAlive() calls.
type fakeAPIClient struct {
	client.APIClient
}

func (f *fakeAPIClient) ContainerInspect(_ context.Context, _ string) (container.InspectResponse, error) {
	resp := container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			State:      &container.State{Status: container.StateRunning},
			HostConfig: &container.HostConfig{},
		},
		Config: &container.Config{Image: "img"},
	}

	return resp, nil
}

// Test_TaskV2_ConcurrentIsAliveAndGetPortBinding_NoRace proves the
// stateMu fix on TaskV2.containerState actually closes the race: a single
// long-lived TaskV2 (as used by configuration.SharedInstance, reused across
// several parallel SetupMatreshka callers) has IsAlive() called from a
// background keep-alive ticker while GetPortBinding() is called from
// multiple test environments concurrently. Run with -race.
func Test_TaskV2_ConcurrentIsAliveAndGetPortBinding_NoRace(t *testing.T) {
	task := &TaskV2{
		container: container.CreateRequest{
			Config: &container.Config{Hostname: "matreshka", Image: "img"},
		},
		dockerAPI: &fakeAPIClient{},
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()

			task.IsAlive()
		}()
		go func() {
			defer wg.Done()

			task.GetPortBinding("50049")
		}()
	}

	wg.Wait()
}
