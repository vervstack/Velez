package network_owner_test

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/network_owner"
)

const (
	rootId = "aaa111"
	midId  = "bbb222"
)

func newSummary(id, name, networkMode string) container.Summary {
	summary := container.Summary{
		ID:    id,
		Names: []string{"/" + name},
	}

	summary.HostConfig.NetworkMode = networkMode

	return summary
}

func Test_OwnerIds_Resolution(t *testing.T) {
	cases := []struct {
		name string
		list []container.Summary
		want map[string]string
	}{
		{
			name: "no sharing",
			list: []container.Summary{
				newSummary(rootId, "app", "bridge"),
				newSummary("bbb222", "db", "host"),
			},
			want: map[string]string{},
		},
		{
			name: "one sidecar by name",
			list: []container.Summary{
				newSummary(rootId, "app", "bridge"),
				newSummary(midId, "ts", "container:app"),
			},
			want: map[string]string{midId: rootId},
		},
		{
			name: "one sidecar by id prefix",
			list: []container.Summary{
				newSummary(rootId, "app", "bridge"),
				newSummary(midId, "ts", "container:aaa1"),
			},
			want: map[string]string{midId: rootId},
		},
		{
			name: "chain resolves to root",
			list: []container.Summary{
				newSummary(rootId, "root", "bridge"),
				newSummary(midId, "mid", "container:root"),
				newSummary("ccc333", "leaf", "container:mid"),
			},
			want: map[string]string{"bbb222": "aaa111", "ccc333": rootId},
		},
		{
			name: "cycle is absent",
			list: []container.Summary{
				newSummary("aaa111", "a", "container:b"),
				newSummary("bbb222", "b", "container:a"),
			},
			want: map[string]string{},
		},
		{
			name: "dangling target is absent",
			list: []container.Summary{
				newSummary("aaa111", "a", "container:missing"),
			},
			want: map[string]string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := network_owner.OwnerIds(tc.list)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_SidecarsOf_ReturnsSharersInListOrder(t *testing.T) {
	list := []container.Summary{
		newSummary("ccc333", "leaf", "container:mid"),
		newSummary(rootId, "root", "bridge"),
		newSummary(midId, "mid", "container:root"),
		newSummary("ddd444", "other", "bridge"),
	}

	got := network_owner.SidecarsOf(list, rootId)

	require.Len(t, got, 2)
	require.Equal(t, "ccc333", got[0].ID)
	require.Equal(t, "bbb222", got[1].ID)
}
