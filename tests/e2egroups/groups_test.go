package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func listSuiteTests(t *testing.T) []string {
	t.Helper()

	root, err := findRepoRoot()
	require.NoError(t, err)

	files, err := filepath.Glob(filepath.Join(root, e2ePackageDir, "*_test.go"))
	require.NoError(t, err)

	var names []string

	for _, file := range files {
		parsed, parseErr := parser.ParseFile(token.NewFileSet(), file, nil, parser.SkipObjectResolution)
		require.NoError(t, parseErr)

		for _, decl := range parsed.Decls {
			fn, isFunc := decl.(*ast.FuncDecl)
			if !isFunc || fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "Test") || fn.Name.Name == "TestMain" {
				continue
			}

			names = append(names, fn.Name.Name)
		}
	}

	return names
}

func Test_GroupsFile_PartitionsEveryTest(t *testing.T) {
	groups, err := loadGroups()
	require.NoError(t, err)

	tests := listSuiteTests(t)
	require.NotEmpty(t, tests)

	err = checkPartition(groups, tests)
	require.NoError(t, err)
}
