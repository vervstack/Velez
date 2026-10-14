package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.redsock.ru/rerrors"
)

const (
	e2ePackageDir = "tests/e2e"
	buildTags     = "e2e_full"
	binaryName    = "e2e.test"
)

// compileSuite builds the e2e test binary once; every group process reuses it.
func compileSuite(ctx context.Context, root, outDir string) (string, error) {
	bin := filepath.Join(outDir, binaryName)

	cmd := exec.CommandContext(ctx, "go", "test", "-tags", buildTags, "-c", "-o", bin, "./"+e2ePackageDir+"/")

	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", rerrors.Wrap(err, "error compiling e2e test binary")
	}

	return bin, nil
}

// listTests returns the top-level test names compiled into the binary.
func listTests(ctx context.Context, bin, root string) ([]string, error) {
	cmd := exec.CommandContext(ctx, bin, "-test.list", ".*")

	cmd.Dir = filepath.Join(root, e2ePackageDir)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing tests: "+stderr.String())
	}

	var names []string

	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.HasPrefix(line, "Test") {
			names = append(names, strings.TrimSpace(line))
		}
	}

	return names, nil
}

// buildSuite compiles the binary and verifies the group partition against it.
func buildSuite(ctx context.Context, groups []group, outDir string) (bin, root string, err error) {
	root, err = findRepoRoot()
	if err != nil {
		return "", "", rerrors.Wrap(err, "error finding repo root")
	}

	bin, err = compileSuite(ctx, root, outDir)
	if err != nil {
		return "", "", rerrors.Wrap(err, "error building suite")
	}

	tests, err := listTests(ctx, bin, root)
	if err != nil {
		return "", "", rerrors.Wrap(err, "error listing suite tests")
	}

	err = checkPartition(groups, tests)
	if err != nil {
		return "", "", rerrors.Wrap(err, "error checking group partition")
	}

	return bin, root, nil
}
