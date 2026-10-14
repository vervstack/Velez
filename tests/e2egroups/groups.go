package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.redsock.ru/rerrors"
	"gopkg.in/yaml.v3"
)

const (
	groupsFileRelPath = "tests/e2e/groups.yaml"
	restGroupName     = "rest"
	daemonNamePrefix  = "e2e-"
	containerPrefix   = "velez-dind-"
	cacheVolumeSuffix = "-cache"
)

var (
	//nolint:forbidigo // package-private tool sentinel, not shared/user-facing
	errRepoRootNotFound = rerrors.New("go.mod not found in any parent directory")
	//nolint:forbidigo // package-private tool sentinel, not shared/user-facing
	errGroupInvalid = rerrors.New("group needs a unique name other than rest and a run regex")
	//nolint:forbidigo // package-private tool sentinel, not shared/user-facing
	errTestInManyGroups = rerrors.New("test matches more than one group")
	//nolint:forbidigo // package-private tool sentinel, not shared/user-facing
	errGroupMatchesNothing = rerrors.New("group regex matches no test")
	//nolint:forbidigo // package-private tool sentinel, not shared/user-facing
	errGroupUnknown = rerrors.New("unknown group")
)

type group struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

type groupsFile struct {
	Groups []group `yaml:"groups"`
}

func (g group) daemonName() string {
	return daemonNamePrefix + g.Name
}

func (g group) containerName() string {
	return containerPrefix + g.daemonName()
}

func (g group) cacheVolumeName() string {
	return g.containerName() + cacheVolumeSuffix
}

func (g group) isRest() bool {
	return g.Name == restGroupName
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", rerrors.Wrap(err, "error getting working directory")
	}

	for {
		_, err = os.Stat(filepath.Join(dir, "go.mod"))
		if err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", rerrors.Wrap(errRepoRootNotFound)
		}

		dir = parent
	}
}

// loadGroups returns the explicit groups from groups.yaml followed by the
// implicit rest group (its Run is empty: it is computed against the others).
func loadGroups() ([]group, error) {
	root, err := findRepoRoot()
	if err != nil {
		return nil, rerrors.Wrap(err, "error finding repo root")
	}

	raw, err := os.ReadFile(filepath.Join(root, groupsFileRelPath))
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading groups file")
	}

	var parsed groupsFile

	err = yaml.Unmarshal(raw, &parsed)
	if err != nil {
		return nil, rerrors.Wrap(err, "error parsing groups file")
	}

	err = validateGroups(parsed.Groups)
	if err != nil {
		return nil, rerrors.Wrap(err, "error validating groups")
	}

	rest := group{Name: restGroupName}

	return append(parsed.Groups, rest), nil
}

func validateGroups(groups []group) error {
	seen := make(map[string]struct{}, len(groups))

	for _, g := range groups {
		_, isDuplicate := seen[g.Name]
		if g.Name == "" || g.Name == restGroupName || g.Run == "" || isDuplicate {
			return rerrors.Wrap(errGroupInvalid, g.Name)
		}

		seen[g.Name] = struct{}{}
	}

	return nil
}

// selectGroups returns every group for an empty name, else just the named one.
func selectGroups(groups []group, name string) ([]group, error) {
	if name == "" {
		return groups, nil
	}

	names := make([]string, 0, len(groups))

	for _, g := range groups {
		if g.Name == name {
			return []group{g}, nil
		}

		names = append(names, g.Name)
	}

	return nil, rerrors.Wrap(errGroupUnknown, name+", valid names: "+strings.Join(names, ", "))
}

func explicitGroups(groups []group) []group {
	explicit := make([]group, 0, len(groups))

	for _, g := range groups {
		if !g.isRest() {
			explicit = append(explicit, g)
		}
	}

	return explicit
}

// testArgs are the test-binary flags selecting the group's tests.
func testArgs(g group, groups []group) []string {
	if !g.isRest() {
		return []string{"-test.run", g.Run}
	}

	explicit := explicitGroups(groups)
	runs := make([]string, 0, len(explicit))

	for _, e := range explicit {
		runs = append(runs, e.Run)
	}

	return []string{"-test.skip", strings.Join(runs, "|")}
}

// checkPartition fails when a test belongs to two explicit groups or an
// explicit group's regex selects nothing.
func checkPartition(groups []group, tests []string) error {
	explicit := explicitGroups(groups)
	hits := make(map[string]int, len(explicit))

	var overlaps []string

	for _, name := range tests {
		matched := 0

		for _, g := range explicit {
			re, err := regexp.Compile(g.Run)
			if err != nil {
				return rerrors.Wrap(err, "error compiling regex of group "+g.Name)
			}

			if re.MatchString(name) {
				matched++

				hits[g.Name]++
			}
		}

		if matched > 1 {
			overlaps = append(overlaps, name)
		}
	}

	if len(overlaps) > 0 {
		return rerrors.Wrap(errTestInManyGroups, strings.Join(overlaps, ", "))
	}

	return firstEmptyGroup(explicit, hits)
}

func firstEmptyGroup(explicit []group, hits map[string]int) error {
	for _, g := range explicit {
		if hits[g.Name] == 0 {
			return rerrors.Wrap(errGroupMatchesNothing, g.Name)
		}
	}

	return nil
}
