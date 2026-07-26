package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"google.golang.org/protobuf/proto"
)

const (
	testLabelPrefix      = "test__"
	integrationTestLabel = testLabelPrefix + "integration"
	testCaseNameLabel    = testLabelPrefix + "name"
	minPortToExposeTo    = uint32(18501)
)

const (
	labelValueTrue  = "true"
	labelValueFalse = "false"
)

const (
	HelloWorldAppImage = "godverv/hello_world:v0.0.14"
	PostgresImage      = "postgres:16"
)

func GetServiceName(t *testing.T) string {
	t.Helper()

	return "e2e_" + strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
}

func GetExpectedLabels(t *testing.T) map[string]string {
	t.Helper()

	return map[string]string{
		labels.CreatedWithVelezLabel: labelValueTrue,
		labels.MatreshkaConfigLabel:  labelValueFalse,
		labels.ComposeGroupLabel:     GetServiceName(t),
	}
}

func AssertSmerds(t *testing.T, expected, actual *velez_api.Smerd) {
	t.Helper()

	require.NotEmpty(t, actual.Uuid)
	actual.Uuid = ""

	require.NotEmpty(t, actual.CreatedAt)
	actual.CreatedAt = nil

	require.Len(t, actual.Ports, len(expected.Ports))

	for idx, port := range actual.Ports {
		require.NotNil(t, port.ExposedTo)
		require.GreaterOrEqual(t, *port.ExposedTo, minPortToExposeTo)

		expected.Ports[idx].ExposedTo = port.ExposedTo
	}

	if !proto.Equal(expected, actual) {
		require.Equal(t, expected, actual)
	}
}

func addTestLabels(t *testing.T, m map[string]string) {
	t.Helper()

	m[integrationTestLabel] = labelValueTrue
	m[testCaseNameLabel] = t.Name()
}

func removeTestLabels(m map[string]string) {
	for k := range m {
		if strings.HasPrefix(k, testLabelPrefix) {
			delete(m, k)
		}
	}
}
