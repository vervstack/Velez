package e2e

import (
	"os"
	"testing"

	"github.com/rs/zerolog/log"
)

// TestMain's only job is to guarantee the shared matreshka fixture (if it
// was ever created — see shared_matreshka.go's getSharedMatreshka) is torn
// down exactly once, after every test in this package has finished,
// regardless of which subtest happened to create it first.
func TestMain(m *testing.M) {
	code := m.Run()

	if sharedMatreshka != nil {
		err := sharedMatreshka.Stop()
		if err != nil {
			log.Error().Err(err).Msg("error stopping shared matreshka instance")
		}
	}

	os.Exit(code)
}
