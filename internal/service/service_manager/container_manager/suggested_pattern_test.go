package container_manager

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func Test_SuggestedPattern_Cases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		image string
		want  velez_api.ServicePattern
	}{
		{"postgres:16", velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES},
		{"docker.io/library/postgres:15-alpine", velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES},
		{"my.registry:5000/postgres", velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES},
		{"postgres@sha256:abc", velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES},
		{"registry:2", velez_api.ServicePattern_SERVICE_PATTERN_REGISTRY},
		{"gitlab/gitlab-runner:latest", velez_api.ServicePattern_SERVICE_PATTERN_GITLAB_RUNNER},
		{"ghcr.io/actions/actions-runner:2.300", velez_api.ServicePattern_SERVICE_PATTERN_GITHUB_RUNNER},
		{"joxit/docker-registry-ui:latest", velez_api.ServicePattern_SERVICE_PATTERN_UNSPECIFIED},
		{"nginx", velez_api.ServicePattern_SERVICE_PATTERN_UNSPECIFIED},
		{"", velez_api.ServicePattern_SERVICE_PATTERN_UNSPECIFIED},
	}

	for _, tc := range cases {
		t.Run(tc.image, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, suggestedPattern(tc.image))
		})
	}
}
