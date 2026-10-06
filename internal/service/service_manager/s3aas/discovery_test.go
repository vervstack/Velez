package s3aas

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/cluster/env"
)

type endpointCase struct {
	name       string
	remoteHost string
	port       uint32
	want       string
}

func Test_S3Endpoint_ResolvesHost(t *testing.T) {
	cases := []endpointCase{
		{name: "remote host", remoteHost: "10.0.0.7", port: 3900, want: "http://10.0.0.7:3900"},
		{name: "remote hostname", remoteHost: "node-2.lan", port: 3901, want: "http://node-2.lan:3901"},
	}

	if !env.IsInContainer() {
		cases = append(cases, endpointCase{name: "local bare binary", port: 3900, want: "http://localhost:3900"})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := s3Endpoint("s3-main", tc.remoteHost, tc.port)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_WebUiEndpoint_ResolvesHost(t *testing.T) {
	cases := []endpointCase{
		{name: "remote host", remoteHost: "10.0.0.7", port: 3902, want: "http://10.0.0.7:3902"},
		{name: "remote hostname", remoteHost: "node-2.lan", port: 3903, want: "http://node-2.lan:3903"},
	}

	if !env.IsInContainer() {
		cases = append(cases, endpointCase{name: "local bare binary", port: 3902, want: "http://localhost:3902"})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := webUiEndpoint("s3-main-webui", tc.remoteHost, tc.port)
			require.Equal(t, tc.want, got)
		})
	}
}
