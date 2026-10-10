package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
)

func Test_ClassifyService(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		serviceName  string
		resourceType string
		isDind       bool
		want         []string
	}{
		{"resource wins", "pgaas_db", "pg", true, []string{domain.ResourceLabel("pg")}},
		{"gitlab runner", "gitlab_runner_x", "", false, []string{domain.LabelServiceRunnerGitlab}},
		{"pgaas", "pgaas_db", "", false, []string{domain.LabelServicePgaas}},
		{"s3", "s3_files", "", false, []string{domain.LabelServiceS3}},
		{"dind", "builder", "", true, []string{domain.LabelServiceDind}},
		{"core", "matreshka", "", false, []string{domain.LabelServiceCore}},
		{"app", "my-app", "", false, []string{domain.LabelServiceApp}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := domain.ClassifyService(tc.serviceName, tc.resourceType, tc.isDind)
			require.Equal(t, tc.want, got)
		})
	}
}
