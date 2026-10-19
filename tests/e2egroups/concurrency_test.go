package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_MaxGroupsForMemory(t *testing.T) {
	const gib = 1 << 30

	cases := []struct {
		name       string
		memTotal   int64
		groupCount int
		want       int
	}{
		{"four gib fits two", 4 * gib, 9, 2},
		{"tiny memory still runs one", gib / 2, 9, 1},
		{"large memory is capped by group count", 64 * gib, 9, 9},
		{"exact multiple", 6 * gib, 9, 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, maxGroupsForMemory(tc.memTotal, tc.groupCount))
		})
	}
}

func Test_ParseMaxGroups(t *testing.T) {
	got, err := parseMaxGroups(" 3 ")
	require.NoError(t, err)
	require.Equal(t, 3, got)

	for _, raw := range []string{"0", "-1", "two", ""} {
		_, err = parseMaxGroups(raw)
		require.ErrorIs(t, err, errMaxGroupsInvalid)
	}
}

func Test_RunConcurrently_RespectsLimitAndRunsEveryGroup(t *testing.T) {
	const limit = 2

	groups := []group{{Name: "a"}, {Name: "b"}, {Name: "c"}, {Name: "d"}, {Name: "e"}}

	var (
		running, peak atomic.Int32
		mu            sync.Mutex
		started       []string
	)

	job := func(g group) result {
		now := running.Add(1)

		for {
			old := peak.Load()
			if now <= old || peak.CompareAndSwap(old, now) {
				break
			}
		}

		mu.Lock()

		started = append(started, g.Name)

		mu.Unlock()

		time.Sleep(20 * time.Millisecond)
		running.Add(-1)

		return result{group: g}
	}

	results := runConcurrently(groups, limit, job)

	require.Len(t, results, len(groups))
	require.Equal(t, int32(limit), peak.Load())
	require.ElementsMatch(t, []string{"a", "b", "c", "d", "e"}, started)
}
