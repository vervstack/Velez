package e2e

import (
	"context"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/clients/matreshka/pkg/matreshka_api"
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
	errInitSharedMatreshka  error

	errMatreshkaNotServing     = rerrors.New("shared matreshka is not serving gRPC")
	errMatreshkaVersionUnknown = rerrors.New("go list reported no matreshka module version")

	matreshkaImage = sync.OnceValues(resolveMatreshkaImage)
)

// resolveMatreshkaImage pins the matreshka image to the module version go.mod
// resolves. The product derives it from the binary's build info, which test
// binaries are built without, so the harness has to supply it.
func resolveMatreshkaImage() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), matreshkaVersionTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "list", "-m", "-f", "{{.Version}}", matreshkaModulePath)

	out, err := cmd.Output()
	if err != nil {
		return "", rerrors.Wrap(err, "error resolving matreshka module version")
	}

	moduleVersion := strings.TrimSpace(string(out))
	if moduleVersion == "" {
		return "", rerrors.Wrap(errMatreshkaVersionUnknown)
	}

	return configuration.ImageRepo + ":" + moduleVersion, nil
}

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

		cfg, errInitSharedMatreshka = config.Load(defaultConfigPath)
		if errInitSharedMatreshka != nil {
			return
		}

		// Bind matreshka's gRPC port to the fixed DinD-side port the
		// harness publishes to the bootstrap host (see dind_ports.go /
		// WithMatreshka), instead of the config default.
		cfg.Environment.MatreshkaPort = dindMatreshkaPort

		cfg.Environment.MatreshkaImage, errInitSharedMatreshka = matreshkaImage()
		if errInitSharedMatreshka != nil {
			return
		}

		var nc node_clients.NodeClients

		nc, errInitSharedMatreshka = node_clients.NewNodeClients(ctx, cfg)
		if errInitSharedMatreshka != nil {
			return
		}

		sharedMatreshka, errInitSharedMatreshka = configuration.StartSharedInstance(ctx, cfg, nc)
		if errInitSharedMatreshka != nil {
			return
		}

		errInitSharedMatreshka = waitMatreshkaServes(ctx)
	})

	require.NoError(t, errInitSharedMatreshka)

	return sharedMatreshka
}

const (
	matreshkaReadyTimeout   = 2 * time.Minute
	matreshkaReadyPoll      = time.Second
	matreshkaVersionTimeout = 30 * time.Second

	matreshkaModulePath = "go.vervstack.ru/matreshka"
)

// waitMatreshkaServes holds the fixture until matreshka's gRPC server answers
// through the port the DinD publishes: until then the published port accepts
// the TCP connection and drops it, which clients see as a failed preface.
func waitMatreshkaServes(ctx context.Context) error {
	addr, ok := sharedDind.Addr(dindMatreshkaPort)
	if !ok {
		return rerrors.Wrap(errMatreshkaNotServing, "dind did not publish the matreshka port")
	}

	conn, err := grpc.NewClient("passthrough:///"+addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return rerrors.Wrap(err, "error dialing shared matreshka")
	}

	defer func() { _ = conn.Close() }()

	client := matreshka_api.NewMatreshkaApiClient(conn)

	ctx, cancel := context.WithTimeout(ctx, matreshkaReadyTimeout)
	defer cancel()

	for {
		_, versionErr := client.Version(ctx, &matreshka_api.Version_Request{})
		if status.Code(versionErr) != codes.Unavailable {
			return nil
		}

		select {
		case <-ctx.Done():
			return rerrors.Wrap(errMatreshkaNotServing)
		case <-time.After(matreshkaReadyPoll):
		}
	}
}
