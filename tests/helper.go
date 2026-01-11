package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

const (
	testLabelPrefix      = "test__"
	integrationTestLabel = testLabelPrefix + "integration"
	testCaseNameLabel    = testLabelPrefix + "name"
	minPortToExposeTo    = uint32(18501)
)

const (
	HelloWorldAppImage = "godverv/hello_world:v0.0.14"
	PostgresImage      = "postgres:16"
)

func GetServiceName(t *testing.T) string {
	return strings.ReplaceAll(t.Name(), "/", "_")
}

func GetExpectedLabels(t *testing.T) map[string]string {
	return map[string]string{
		labels.CreatedWithVelezLabel: "true",
		labels.MatreshkaConfigLabel:  "false",
		labels.ComposeGroupLabel:     GetServiceName(t),
	}
}

func AssertSmerds(t *testing.T, expected, actual *velez_api.Smerd) {
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
	m[integrationTestLabel] = "true"
	m[testCaseNameLabel] = t.Name()
}

func removeTestLabels(m map[string]string) {
	for k := range m {
		if strings.HasPrefix(k, testLabelPrefix) {
			delete(m, k)
		}
	}
}
