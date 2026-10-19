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

func Test_ParseGroups_ParallelDefaultsAndOverrides(t *testing.T) {
	raw := "groups:\n  - name: a\n    run: \"^A$\"\n  - name: b\n    run: \"^B$\"\n    parallel: 2\n"

	groups, err := parseGroups([]byte(raw))
	require.NoError(t, err)
	require.Len(t, groups, 3)

	require.Equal(t, "4", groups[0].testParallel())
	require.Equal(t, "2", groups[1].testParallel())
	require.Equal(t, "4", groups[2].testParallel())
}

func Test_ParseGroups_RejectsNonPositiveParallel(t *testing.T) {
	for _, value := range []string{"0", "-1"} {
		raw := "groups:\n  - name: a\n    run: \"^A$\"\n    parallel: " + value + "\n"

		_, err := parseGroups([]byte(raw))
		require.ErrorIs(t, err, errParallelInvalid, value)
	}
}

func Test_RunArgs_RunnersGroupRunsTwoTestsAtATime(t *testing.T) {
	groups, err := loadGroups()
	require.NoError(t, err)

	selected, err := selectGroups(groups, "runners")
	require.NoError(t, err)

	args := runArgs(selected[0], groups, "")
	require.Contains(t, strings.Join(args, " "), "-test.parallel 2 ")

	rest, err := selectGroups(groups, restGroupName)
	require.NoError(t, err)

	args = runArgs(rest[0], groups, "")
	require.Contains(t, strings.Join(args, " "), "-test.parallel 4 ")
}
