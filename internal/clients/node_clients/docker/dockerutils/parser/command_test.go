package parser

import (
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func garageStatusExec() []string {
	return []string{"/garage", "status"}
}

func Test_FromHealthcheck_Cases(t *testing.T) {
	t.Parallel()

	command := "pg_isready"

	cases := []struct {
		name     string
		in       *velez_api.Container_Healthcheck
		wantTest []string
		isNil    bool
	}{
		{
			name:  "nil healthcheck",
			in:    nil,
			isNil: true,
		},
		{
			name:     "exec form",
			in:       &velez_api.Container_Healthcheck{Exec: garageStatusExec()},
			wantTest: append([]string{healthcheckTestExec}, garageStatusExec()...),
		},
		{
			name:     "command form",
			in:       &velez_api.Container_Healthcheck{Command: &command},
			wantTest: []string{healthcheckTestShell, "pg_isready"},
		},
		{
			name:     "exec wins over command",
			in:       &velez_api.Container_Healthcheck{Command: &command, Exec: []string{"/bin/true"}},
			wantTest: []string{healthcheckTestExec, "/bin/true"},
		},
		{
			name:  "neither command nor exec",
			in:    &velez_api.Container_Healthcheck{IntervalSecond: 2, Retries: 3},
			isNil: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FromHealthcheck(tc.in)
			if tc.isNil {
				require.Nil(t, got)

				return
			}

			require.NotNil(t, got)
			require.Equal(t, tc.wantTest, got.Test)
		})
	}
}

func Test_FromHealthcheck_CopiesTimings(t *testing.T) {
	t.Parallel()

	timeout := uint32(5)
	in := &velez_api.Container_Healthcheck{
		Exec:           garageStatusExec(),
		IntervalSecond: 2,
		TimeoutSecond:  &timeout,
		Retries:        30,
	}

	got := FromHealthcheck(in)

	require.Equal(t, 2*time.Second, got.Interval)
	require.Equal(t, 5*time.Second, got.Timeout)
	require.Equal(t, 30, got.Retries)
}

func Test_ToHealthcheck_ReversesExecForm(t *testing.T) {
	t.Parallel()

	got := ToHealthcheck(&container.HealthConfig{Test: append([]string{healthcheckTestExec}, garageStatusExec()...)})

	require.Equal(t, garageStatusExec(), got.GetExec())
	require.Nil(t, got.Command)
}
