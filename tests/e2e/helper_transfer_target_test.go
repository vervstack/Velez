//go:build e2e_full

package e2e

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/tests/dind"
)

const transferTargetDindSuffix = "-target"

var (
	transferTargetOnce sync.Once //nolint:gochecknoglobals // one target daemon per test binary
	transferTargetErr  error     //nolint:gochecknoglobals // outcome of the one-time bring-up
)

// transferTargetDockerHost brings up, on first use, the second DinD daemon the
// transfer target Velez runs against, so a moved container really leaves the
// source daemon's name and volume namespace.
func transferTargetDockerHost(t *testing.T) string {
	t.Helper()

	transferTargetOnce.Do(func() {
		opts := dind.Options{
			BootstrapHost: sharedDind.BootstrapHost,
			NameSuffix:    transferTargetDindSuffix,
		}

		env, err := dind.Setup(context.Background(), opts)
		if err != nil {
			transferTargetErr = err

			return
		}

		sharedTargetDind = env

		transferTargetErr = env.Reset(context.Background())
	})

	require.NoError(t, transferTargetErr)

	return sharedTargetDind.DockerHost
}
