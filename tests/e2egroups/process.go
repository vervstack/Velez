package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"go.redsock.ru/rerrors"
)

const (
	extraEnvCount   = 3
	envLogDir       = "VELEZ_E2E_LOG_DIR"
	envDindName     = "VELEZ_E2E_DIND_NAME"
	envPrepareOnly  = "VELEZ_E2E_PREPARE"
	shutdownTimeout = 3 * time.Minute
)

type result struct {
	group    group
	logPath  string
	duration time.Duration
	err      error
}

func (r result) isPassed() bool {
	return r.err == nil
}

// groupEnv is the inherited environment pinned to the group's daemon.
func groupEnv(g group, isPrepare bool) []string {
	env := make([]string, 0, len(os.Environ())+extraEnvCount)

	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, envPrepareOnly+"=") || strings.HasPrefix(kv, envDindName+"=") ||
			strings.HasPrefix(kv, envVelezImage+"=") {
			continue
		}

		env = append(env, kv)
	}

	env = append(env, envDindName+"="+g.daemonName(), envVelezImage+"="+velezImageRef)

	if isPrepare {
		env = append(env, envPrepareOnly+"=1")
	}

	return env
}

// runProcess runs the test binary for one group, output going to logPath. A
// cancelled ctx sends SIGTERM so the child's TestMain tears its daemon down.
func runProcess(ctx context.Context, bin, root, logPath string, g group, isPrepare bool, args []string) result {
	started := time.Now()
	res := result{group: g, logPath: logPath}

	logFile, err := os.Create(logPath)
	if err != nil {
		res.err = rerrors.Wrap(err, "error creating group log")

		return res
	}

	defer func() { _ = logFile.Close() }()

	cmd := exec.CommandContext(ctx, bin, args...)

	cmd.Dir = filepath.Join(root, e2ePackageDir)
	cmd.Env = groupEnv(g, isPrepare)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = shutdownTimeout

	err = cmd.Run()
	if err != nil {
		res.err = rerrors.Wrap(err, "error running group "+g.Name)
	}

	res.duration = time.Since(started)

	return res
}

// runConcurrently runs at most limit jobs at once and prints a line as each ends.
func runConcurrently(groups []group, limit int, job func(g group) result) []result {
	results := make(chan result, len(groups))
	slots := make(chan struct{}, limit)

	for _, g := range groups {
		go func() {
			slots <- struct{}{}

			defer func() { <-slots }()

			results <- job(g)
		}()
	}

	collected := make([]result, 0, len(groups))

	for range groups {
		res := <-results

		printResult(res)

		collected = append(collected, res)
	}

	return collected
}
