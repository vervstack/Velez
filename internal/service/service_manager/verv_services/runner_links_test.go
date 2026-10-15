package verv_services

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	squidName = "squid"
)

func Test_ProxyHostOf_Cases(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"http with port", "http://squid:3128", squidName},
		{"socks5 ip", "socks5://10.0.0.5:1080", "10.0.0.5"},
		{"not a url", "not a url", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, proxyHostOf(tc.input))
		})
	}
}

func Test_FindManagedContainer_Cases(t *testing.T) {
	cases := []struct {
		name       string
		containers []container.Summary
		search     string
		wantFound  bool
		wantId     string
	}{
		{
			name: "matches service label",
			containers: []container.Summary{
				{ID: "a", Labels: map[string]string{labels.VervServiceLabel: squidName}},
			},
			search:    squidName,
			wantFound: true,
			wantId:    "a",
		},
		{
			name: "matches container name",
			containers: []container.Summary{
				{ID: "b", Names: []string{"/squid"}, Labels: map[string]string{labels.VervServiceLabel: "proxy-svc"}},
			},
			search:    squidName,
			wantFound: true,
			wantId:    "b",
		},
		{
			name: "skips sidecar",
			containers: []container.Summary{
				{
					ID: "c",
					Labels: map[string]string{
						labels.VervServiceLabel: squidName,
						labels.Sidecar:          "true",
					},
				},
			},
			search:    squidName,
			wantFound: false,
		},
		{
			name: "container without service label is not managed",
			containers: []container.Summary{
				{ID: "d", Names: []string{"/squid"}},
			},
			search:    squidName,
			wantFound: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, isFound := findManagedContainer(tc.containers, tc.search)
			require.Equal(t, tc.wantFound, isFound)
			require.Equal(t, tc.wantId, got.ID)
		})
	}
}
