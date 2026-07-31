package env

import (
	"context"
	"sync"
	"testing"

	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// fakeVolumeAPIClient embeds client.APIClient as a nil interface value
// purely so this type satisfies the interface without minimock; only
// VolumeList/VolumeCreate are overridden, since that's all StartVolumes calls.
type fakeVolumeAPIClient struct {
	client.APIClient
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

	for range 20 {
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

// Test_StartVolumes_Repeated_NameDoesNotCompound guards against a regression
// where vervVolumeName was mutated with `+=` on every call instead of being
// recomputed from a stable base, causing repeated StartVolumes calls (e.g.
// service restarts) to produce ever-growing names like
// "verv_host_host_host..." and orphan a new Docker volume each time.
func Test_StartVolumes_Repeated_NameDoesNotCompound(t *testing.T) {
	dockerAPI := &fakeVolumeAPIClient{}

	err := StartVolumes(dockerAPI)
	if err != nil {
		t.Fatal(err)
	}

	firstName := GetVervVolumeName()

	for range 5 {
		err = StartVolumes(dockerAPI)
		if err != nil {
			t.Fatal(err)
		}
	}

	secondName := GetVervVolumeName()

	if firstName != secondName {
		t.Errorf("expected volume name to stay %q after repeated StartVolumes calls, got %q", firstName, secondName)
	}
}
