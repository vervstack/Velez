package pgaas_api_impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func Test_RequestHostname_Sources(t *testing.T) {
	cases := []struct {
		name   string
		header map[string]string
		want   string
	}{
		{"forwarded host with port", map[string]string{forwardedHostHeader: "velez.example.com:53890"}, "velez.example.com"},
		{"forwarded host without port", map[string]string{forwardedHostHeader: "10.0.0.5"}, "10.0.0.5"},
		{"authority fallback", map[string]string{authorityHeader: "node.local:53890"}, "node.local"},
		{"forwarded host wins over authority", map[string]string{
			forwardedHostHeader: "public.host", authorityHeader: "internal:53890",
		}, "public.host"},
		{"no metadata", nil, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if tc.header != nil {
				ctx = metadata.NewIncomingContext(ctx, metadata.New(tc.header))
			}

			require.Equal(t, tc.want, requestHostname(ctx))
		})
	}
}
