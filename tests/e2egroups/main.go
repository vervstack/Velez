// Command e2egroups runs the e2e suite as several test processes at once,
// one per group in tests/e2e/groups.yaml, each on its own prepared DinD
// daemon. See "Full parallel run" in tests/e2e/CLAUDE.md.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.redsock.ru/rerrors"
)

const (
	commandPrepare = "prepare"
	commandRun     = "run"
	commandClean   = "clean"
)

//nolint:forbidigo // package-private tool sentinel, not shared/user-facing
var errUsage = rerrors.New("usage: go run ./tests/e2egroups <prepare|run|clean>")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	err := dispatch(ctx, os.Args[1:])

	stop()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func dispatch(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return rerrors.Wrap(errUsage)
	}

	groups, err := loadGroups()
	if err != nil {
		return rerrors.Wrap(err, "error loading groups")
	}

	switch args[0] {
	case commandPrepare:
		return prepare(ctx, groups)
	case commandRun:
		return runAll(ctx, groups)
	case commandClean:
		return clean(ctx, groups)
	default:
		return rerrors.Wrap(errUsage)
	}
}
