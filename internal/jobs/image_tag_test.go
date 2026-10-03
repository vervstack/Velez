package jobs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ImageWithTag(t *testing.T) {
	digest := "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	cases := []struct {
		name    string
		image   string
		tag     string
		want    string
		isError bool
	}{
		{name: "replaces tag", image: "nginx:latest", tag: "1.27.2", want: "nginx:1.27.2"},
		{name: "adds tag to bare name", image: "nginx", tag: "1.27.2", want: "nginx:1.27.2"},
		{name: "keeps registry host", image: "ghcr.io/o/r:latest", tag: "v1", want: "ghcr.io/o/r:v1"},
		{name: "drops digest", image: "nginx@" + digest, tag: "1.27", want: "nginx:1.27"},
		{name: "invalid tag", image: "nginx:latest", tag: "-bad tag", isError: true},
		{name: "invalid reference", image: "NGINX::", tag: "1", isError: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := imageWithTag(tc.image, tc.tag)
			if tc.isError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
