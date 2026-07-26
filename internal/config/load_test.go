package config

import (
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test_Load_Concurrent_NoRace proves Load holds no shared state: many
// goroutines calling it concurrently (mirroring parallel e2e
// TestEnvironments, each of which calls config.Load(configPath) during
// setup) never race, no matter how many call it at once. Run with -race.
func Test_Load_Concurrent_NoRace(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	// filename -> internal/config/load_test.go; go up two levels to repo root.
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	configPath := filepath.Join(repoRoot, "tests", "config_mocks", "velez_default_config.yaml")

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := Load(configPath)
			require.NoError(t, err)
		}()
	}
	wg.Wait()
}
