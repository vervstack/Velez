package env

import (
	"context"
	"sync"
	"testing"

	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// fakeVolumeAPIClient embeds client.CommonAPIClient as a nil interface value
// purely so this type satisfies the interface without minimock; only
// VolumeList/VolumeCreate are overridden, since that's all StartVolumes calls.
type fakeVolumeAPIClient struct {
	client.CommonAPIClient
}

func (f *fakeVolumeAPIClient) VolumeList(_ context.Context, _ volume.ListOptions) (volume.ListResponse, error) {
	return volume.ListResponse{}, nil
}

func (f *fakeVolumeAPIClient) VolumeCreate(_ context.Context, options volume.CreateOptions) (volume.Volume, error) {
	return volume.Volume{Name: options.Name, Options: options.DriverOpts}, nil
}

// Test_StartVolumes_Concurrent_NoRace proves the volumeMu fix closes the race
// on vervVolumeName/vervVolumePath: multiple in-process environments (e.g.
// parallel e2e TestEnvironments) each call StartVolumes concurrently. Run
// with -race.
func Test_StartVolumes_Concurrent_NoRace(t *testing.T) {
	dockerAPI := &fakeVolumeAPIClient{}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := StartVolumes(dockerAPI)
			if err != nil {
				t.Error(err)
			}

			_ = GetVervVolumeName()
			_, _ = GetVervVolumePath()
		}()
	}
	wg.Wait()
}
