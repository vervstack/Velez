package image_versions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SelectRepositoryAndDigest(t *testing.T) {
	cases := []struct {
		name           string
		configImage    string
		repoDigests    []string
		wantRepository string
		wantDigest     string
		wantResolved   bool
	}{
		{
			name:           "docker hub image with tag",
			configImage:    "nginx:latest",
			repoDigests:    []string{"nginx@sha256:aa"},
			wantRepository: "docker.io/library/nginx",
			wantDigest:     "sha256:aa",
			wantResolved:   true,
		},
		{
			name:           "picks digest of the matching repository",
			configImage:    "ghcr.io/o/r:1",
			repoDigests:    []string{"docker.io/library/other@sha256:bb", "ghcr.io/o/r@sha256:cc"},
			wantRepository: "ghcr.io/o/r",
			wantDigest:     "sha256:cc",
			wantResolved:   true,
		},
		{
			name:           "container run by image id falls back to first repo digest",
			configImage:    "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			repoDigests:    []string{"ghcr.io/o/r@sha256:dd"},
			wantRepository: "ghcr.io/o/r",
			wantDigest:     "sha256:dd",
			wantResolved:   true,
		},
		{
			name:        "no repo digests is not resolvable",
			configImage: "nginx:latest",
			repoDigests: nil,
		},
		{
			name:        "no digest for the config repository is not resolvable",
			configImage: "myreg.local/app:1",
			repoDigests: []string{"nginx@sha256:aa"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repository, digest, isResolved := selectRepositoryAndDigest(tc.configImage, tc.repoDigests)

			require.Equal(t, tc.wantResolved, isResolved)
			require.Equal(t, tc.wantDigest, digest)

			if tc.wantResolved {
				require.Equal(t, tc.wantRepository, repository.Name())
			}
		})
	}
}
