package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"

	"go.redsock.ru/rerrors"
)

const (
	envSocksBin       = "VELEZ_E2E_SOCKS_BIN"
	socksPackage      = "./tests/dind/socksproxy"
	socksBinNamePart  = "velez-e2e-socksproxy-"
	socksPathHashSize = 4
)

// socksBinPath is a per-checkout path, so parallel worktrees never rebuild the
// same file under each other.
func socksBinPath(root string) string {
	sum := sha256.Sum256([]byte(root))

	return filepath.Join(os.TempDir(), socksBinNamePart+hex.EncodeToString(sum[:socksPathHashSize]))
}

// buildSocksProxy cross-compiles the SOCKS5 proxy the harness runs inside
// every DinD, for the same architecture as the velez image.
func buildSocksProxy(ctx context.Context, root, arch string) error {
	goBuild := exec.CommandContext(ctx, "go", "build", "-o", socksBinPath(root), socksPackage)

	goBuild.Dir = root

	goBuild.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+arch, "CGO_ENABLED=0")

	err := runBuildStep(goBuild, "error cross-compiling socks proxy")
	if err != nil {
		return rerrors.Wrap(err, "error building socks proxy binary")
	}

	return nil
}
