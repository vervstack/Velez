package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"go.redsock.ru/rerrors"
)

const (
	logDirPerm      = 0o755
	preparedTestRun = "^$"
	testParallel    = "4"
	testTimeout     = "30m"
)

//nolint:forbidigo // package-private tool sentinel, not shared/user-facing
var (
	errGroupsFailed = rerrors.New("some groups failed")
	errNotPrepared  = rerrors.New("a group has no prepared cache volume, run `make e2e-prepare`")
)

func prepareLogDir() (string, error) {
	dir := os.Getenv(envLogDir)
	if dir == "" {
		tmp, err := os.MkdirTemp("", "velez-e2e-logs-")
		if err != nil {
			return "", rerrors.Wrap(err, "error creating log dir")
		}

		dir = tmp
	}

	err := os.MkdirAll(dir, logDirPerm)
	if err != nil {
		return "", rerrors.Wrap(err, "error creating log dir")
	}

	fmt.Println("logs:", dir)

	return dir, nil
}

func prepare(ctx context.Context, groups []group) error {
	return execute(ctx, groups, func(ctx context.Context, bin, root, logDir string) []result {
		return runConcurrently(groups, func(g group) result {
			logPath := filepath.Join(logDir, g.Name+".prepare.log")
			args := []string{"-test.run", preparedTestRun}

			return runProcess(ctx, bin, root, logPath, g, true, args)
		})
	})
}

func runAll(ctx context.Context, groups []group) error {
	for _, g := range groups {
		err := requireCacheVolume(ctx, g)
		if err != nil {
			return rerrors.Wrap(err, "error checking prepared daemon of group "+g.Name)
		}
	}

	return execute(ctx, groups, func(ctx context.Context, bin, root, logDir string) []result {
		return runConcurrently(groups, func(g group) result {
			logPath := filepath.Join(logDir, g.Name+".log")
			args := append(testArgs(g, groups), "-test.parallel", testParallel, "-test.count=1", "-test.timeout", testTimeout)

			return runProcess(ctx, bin, root, logPath, g, false, args)
		})
	})
}

type groupsJob func(ctx context.Context, bin, root, logDir string) []result

func execute(ctx context.Context, groups []group, job groupsJob) error {
	logDir, err := prepareLogDir()
	if err != nil {
		return rerrors.Wrap(err, "error preparing log dir")
	}

	binDir, err := os.MkdirTemp("", "velez-e2e-bin-")
	if err != nil {
		return rerrors.Wrap(err, "error creating binary dir")
	}

	defer func() { _ = os.RemoveAll(binDir) }()

	bin, root, err := buildSuite(ctx, groups, binDir)
	if err != nil {
		return rerrors.Wrap(err, "error building suite")
	}

	results := job(ctx, bin, root, logDir)

	return summarize(results, logDir)
}

func printResult(res result) {
	status := "PASS"
	if !res.isPassed() {
		status = "FAIL"
	}

	fmt.Printf("%s: %s (%s)\n", res.group.Name, status, res.duration.Round(time.Second))
}

func summarize(results []result, logDir string) error {
	sort.Slice(results, func(i, j int) bool { return results[i].group.Name < results[j].group.Name })

	var failed bool

	for _, res := range results {
		if res.isPassed() {
			continue
		}

		failed = true

		fmt.Printf("\n%s failed (%s):\n%s\n", res.group.Name, res.logPath, failureLines(res.logPath))
	}

	fmt.Println("logs:", logDir)

	if failed {
		return rerrors.Wrap(errGroupsFailed)
	}

	return nil
}

func clean(ctx context.Context, groups []group) error {
	for _, g := range groups {
		err := removeContainer(ctx, g.containerName())
		if err != nil {
			return rerrors.Wrap(err, "error cleaning group "+g.Name)
		}

		err = removeVolume(ctx, g.cacheVolumeName())
		if err != nil {
			return rerrors.Wrap(err, "error cleaning group "+g.Name)
		}

		fmt.Println("cleaned:", g.Name)
	}

	return nil
}
