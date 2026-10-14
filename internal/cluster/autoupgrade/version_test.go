package autoupgrade

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	tag100       = "1.0.0"
	tag102       = "1.0.2"
	tag103       = "1.0.3"
	tag150       = "1.5.0"
	tag200       = "2.0.0"
	tagV103      = "v1.0.3"
	tagLatest    = "latest"
	tagStable    = "stable"
	repoLocal    = "localhost:5000/e2e/app"
	repoNginx    = "nginx"
	repoRegistry = "ghcr.io/a/b"
)

func Test_PickTag_Scenarios(t *testing.T) {
	major := velez_api.AutoUpdateBumpLevel_MAJOR
	minor := velez_api.AutoUpdateBumpLevel_MINOR
	patch := velez_api.AutoUpdateBumpLevel_PATCH
	unspecified := velez_api.AutoUpdateBumpLevel_AUTO_UPDATE_BUMP_LEVEL_UNSPECIFIED

	allBumps := []string{tag103, tag150, tag200}
	noMajor := []string{tag103, tag150}
	patchOnly := []string{tag103}

	cases := []struct {
		name    string
		current string
		tags    []string
		level   velez_api.AutoUpdateBumpLevel
		want    string
		isMove  bool
	}{
		{"major level takes major bump", tag102, allBumps, major, tag200, true},
		{"minor level takes highest accepted", tag102, allBumps, minor, tag200, true},
		{"patch level takes highest accepted", tag102, allBumps, patch, tag200, true},
		{"major level skips minor and patch", tag102, noMajor, major, "", false},
		{"minor level takes minor bump", tag102, noMajor, minor, tag150, true},
		{"patch level takes minor bump", tag102, noMajor, patch, tag150, true},
		{"minor level skips patch only", tag102, patchOnly, minor, "", false},
		{"patch level takes patch bump", tag102, patchOnly, patch, tag103, true},
		{"unspecified level behaves as patch", tag102, patchOnly, unspecified, tag103, true},
		{"compares numerically", tag100, []string{"1.9.0", "1.10.0", "1.2.0"}, patch, "1.10.0", true},
		{"patch numerically beats lexical", tag102, []string{"1.0.9", "1.0.10"}, patch, "1.0.10", true},
		{"ignores pre-release latest and short tags", tag102,
			[]string{tagLatest, "2.0.0-rc1", "2.0", "3", "1.0.3+build", tag103}, patch, tag103, true},
		{"only non-semver candidates means no move", tag102, []string{tagLatest, "2.0.0-rc1", "2.0"}, patch, "", false},
		{"current is already newest", tag103, []string{"1.0.1", tag103}, patch, "", false},
		{"older tags are ignored", tag200, []string{tag100, "1.9.9"}, patch, "", false},
		{"v-prefixed tags keep their original string", "v1.0.2", []string{tagV103, "v1.1.0"}, patch, "v1.1.0", true},
		{"v-prefixed current with bare candidates", "v1.0.2", []string{tag103}, patch, tag103, true},
		{"bare current with v-prefixed candidates", tag102, []string{tagV103}, patch, tagV103, true},
		{"legacy current empty takes last tag", "", []string{"a", "b", tagLatest}, major, tagLatest, true},
		{"legacy current non-semver takes last tag", tagLatest, []string{tag100, tagStable}, major, tagStable, true},
		{"legacy current equal to last tag does not move", tagStable, []string{tag100, tagStable}, patch, "", false},
		{"legacy with no tags does not move", "", nil, patch, "", false},
		{"semver current with no tags does not move", tag102, nil, patch, "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, isMove := pickTag(tc.current, tc.tags, tc.level)

			require.Equal(t, tc.isMove, isMove)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_SplitImage_Scenarios(t *testing.T) {
	cases := []struct {
		name  string
		image string
		want  imageRef
	}{
		{"host with port and tag", "localhost:5000/e2e/app:1.0.0", imageRef{Repo: repoLocal, Tag: tag100}},
		{"host with port without tag", repoLocal, imageRef{Repo: repoLocal}},
		{"plain with tag", "nginx:1.2.3", imageRef{Repo: repoNginx, Tag: "1.2.3"}},
		{"plain without tag", repoNginx, imageRef{Repo: repoNginx}},
		{"registry host with tag", "ghcr.io/a/b:1.0.0", imageRef{Repo: repoRegistry, Tag: tag100}},
		{"registry host without tag", repoRegistry, imageRef{Repo: repoRegistry}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, splitImage(tc.image))
		})
	}
}

func Test_BumpLevelFromLabels_Scenarios(t *testing.T) {
	cases := []struct {
		name   string
		labels map[string]string
		want   velez_api.AutoUpdateBumpLevel
	}{
		{"major", map[string]string{labels.AutoUpgradeBumpLevel: bumpLevelLabelMajor}, velez_api.AutoUpdateBumpLevel_MAJOR},
		{"minor", map[string]string{labels.AutoUpgradeBumpLevel: bumpLevelLabelMinor}, velez_api.AutoUpdateBumpLevel_MINOR},
		{"patch", map[string]string{labels.AutoUpgradeBumpLevel: bumpLevelLabelPatch}, velez_api.AutoUpdateBumpLevel_PATCH},
		{"absent", map[string]string{}, velez_api.AutoUpdateBumpLevel_PATCH},
		{"nil labels", nil, velez_api.AutoUpdateBumpLevel_PATCH},
		{"garbage", map[string]string{labels.AutoUpgradeBumpLevel: "huge"}, velez_api.AutoUpdateBumpLevel_PATCH},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, bumpLevelFromLabels(tc.labels))
		})
	}
}
