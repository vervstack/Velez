// Command e2egroups runs the e2e suite as several test processes at once,
// one per group in tests/e2e/groups.yaml, each on its own prepared DinD
// daemon. See "Full parallel run" in tests/e2e/CLAUDE.md.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"go.redsock.ru/rerrors"
)

const (
	commandPrepare = "prepare"
	commandRun     = "run"
	commandClean   = "clean"
	commandImage   = "image"
)

//nolint:forbidigo // package-private tool sentinel, not shared/user-facing
var errUsage = rerrors.New(
	"usage: go run ./tests/e2egroups " +
		"<image | prepare [--group <name>] | " +
		"run [--group <name>] [--prepare-missing] [--filter <regex>] [--max-groups <n>] | clean>",
)

type options struct {
	groupName        string
	isPrepareMissing bool
	testFilter       string
	maxGroups        int
}

func parseOptions(command string, args []string) (options, error) {
	var opts options

	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	if command != commandClean {
		flags.StringVar(&opts.groupName, "group", "", "limit to one group")
	}

	if command == commandRun {
		flags.BoolVar(&opts.isPrepareMissing, "prepare-missing", false, "prepare groups whose cache volume is absent")
		flags.StringVar(&opts.testFilter, "filter", "", "narrow the selected groups' tests to this -test.run regex")
		flags.IntVar(&opts.maxGroups, "max-groups", 0,
			"most groups running at once (default: "+envMaxGroups+", else from docker memory)")
	}

	err := flags.Parse(args)
	if err != nil {
		return options{}, rerrors.Wrap(errUsage, err.Error())
	}

	if flags.NArg() != 0 {
		return options{}, rerrors.Wrap(errUsage)
	}

	return opts, nil
}

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
	if len(args) == 0 {
		return rerrors.Wrap(errUsage)
	}

	command := args[0]
	if command != commandPrepare && command != commandRun && command != commandClean && command != commandImage {
		return rerrors.Wrap(errUsage)
	}

	if command == commandImage {
		if len(args) != 1 {
			return rerrors.Wrap(errUsage)
		}

		return buildVelezImage(ctx)
	}

	opts, err := parseOptions(command, args[1:])
	if err != nil {
		return rerrors.Wrap(err, "error parsing options")
	}

	groups, err := loadGroups()
	if err != nil {
		return rerrors.Wrap(err, "error loading groups")
	}

	selected, err := selectGroups(groups, opts.groupName)
	if err != nil {
		return rerrors.Wrap(err, "error selecting groups")
	}

	if command != commandClean {
		err = buildVelezImage(ctx)
		if err != nil {
			return rerrors.Wrap(err, "error building velez image")
		}
	}

	switch command {
	case commandPrepare:
		return prepare(ctx, groups, selected)
	case commandRun:
		return runAll(ctx, groups, selected, opts.isPrepareMissing, opts.testFilter, opts.maxGroups)
	default:
		return clean(ctx, groups)
	}
}
