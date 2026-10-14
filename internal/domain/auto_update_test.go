package domain

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

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
			require.Equal(t, tc.want, BumpLevelFromLabels(tc.labels))
		})
	}
}

func Test_BumpLevelLabelValue_Scenarios(t *testing.T) {
	cases := []struct {
		name  string
		level velez_api.AutoUpdateBumpLevel
		want  string
	}{
		{"major", velez_api.AutoUpdateBumpLevel_MAJOR, bumpLevelLabelMajor},
		{"minor", velez_api.AutoUpdateBumpLevel_MINOR, bumpLevelLabelMinor},
		{"patch", velez_api.AutoUpdateBumpLevel_PATCH, bumpLevelLabelPatch},
		{"unspecified", velez_api.AutoUpdateBumpLevel_AUTO_UPDATE_BUMP_LEVEL_UNSPECIFIED, bumpLevelLabelPatch},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, BumpLevelLabelValue(tc.level))
		})
	}
}

func Test_BumpLevel_RoundTrip(t *testing.T) {
	levels := []velez_api.AutoUpdateBumpLevel{
		velez_api.AutoUpdateBumpLevel_MAJOR,
		velez_api.AutoUpdateBumpLevel_MINOR,
		velez_api.AutoUpdateBumpLevel_PATCH,
	}

	for _, level := range levels {
		containerLabels := map[string]string{labels.AutoUpgradeBumpLevel: BumpLevelLabelValue(level)}

		require.Equal(t, level, BumpLevelFromLabels(containerLabels))
	}
}
