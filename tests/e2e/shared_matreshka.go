package e2e

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/configuration"
	"go.vervstack.ru/Velez/internal/config"
)

// sharedMatreshka is the single, process-wide matreshka container fixture
// used by every WithMatreshka() TestEnvironment in this package. See
// WithMatreshka's doc comment for why this must stay confined to package
// e2e. Torn down by TestMain (main_test.go) after every test finishes.
var (
	sharedMatreshka         *configuration.SharedInstance
	initSharedMatreshkaOnce sync.Once
	initSharedMatreshkaErr  error
)

// getSharedMatreshka lazily creates the one matreshka container + keep-alive
// loop shared by every WithMatreshka() TestEnvironment in this package, the
// first time any test calls WithMatreshka(). Mirrors
// tests/test_helper.GetSharedPortManager's existing sync.Once pattern, so
// -run filters that never touch cluster mode don't pay container spin-up
// cost.
func getSharedMatreshka(t *testing.T) *configuration.SharedInstance {
	t.Helper()

	initSharedMatreshkaOnce.Do(func() {
		ctx := context.Background()

		var cfg config.Config
		cfg, initSharedMatreshkaErr = config.Load(defaultConfigPath)
		if initSharedMatreshkaErr != nil {
			return
		}

		var nc node_clients.NodeClients
		nc, initSharedMatreshkaErr = node_clients.NewNodeClients(ctx, cfg)
		if initSharedMatreshkaErr != nil {
			return
		}

		sharedMatreshka, initSharedMatreshkaErr = configuration.StartSharedInstance(ctx, cfg, nc)
	})

	require.NoError(t, initSharedMatreshkaErr)

	return sharedMatreshka
}
