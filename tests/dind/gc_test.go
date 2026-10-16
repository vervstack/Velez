package dind

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_IsCacheStale_Scenarios(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name      string
		lastUsed  time.Time
		isRunning bool
		want      bool
	}{
		{"used 4 days ago is stale", now.Add(-4 * 24 * time.Hour), false, true},
		{"used 2 days ago is kept", now.Add(-2 * 24 * time.Hour), false, false},
		{"running is kept even if old", now.Add(-30 * 24 * time.Hour), true, false},
		{"exactly at the boundary is kept", now.Add(-cacheMaxAge), false, false},
		{"just past the boundary is stale", now.Add(-cacheMaxAge - time.Second), false, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isCacheStale(now, tc.lastUsed, tc.isRunning))
		})
	}
}

func Test_SelectCacheCandidates_Scenarios(t *testing.T) {
	const (
		current      = "velez-dind-master-cache"
		featureCache = "velez-dind-feature-cache"
	)

	cases := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			"other branches are candidates",
			[]string{featureCache, "velez-dind-fix-x-target-cache"},
			[]string{featureCache, "velez-dind-fix-x-target-cache"},
		},
		{
			"current run is skipped",
			[]string{current, featureCache},
			[]string{featureCache},
		},
		{
			"foreign prefix or suffix is skipped",
			[]string{"other-cache", "velez-dind-feature", "postgres-data"},
			[]string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, selectCacheCandidates(tc.input, current))
		})
	}
}

func Test_ContainerNameOfCache_StripsSuffix(t *testing.T) {
	require.Equal(t, "velez-dind-docs", containerNameOfCache("velez-dind-docs-cache"))
}
