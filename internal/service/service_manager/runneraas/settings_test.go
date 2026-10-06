package runneraas

import (
	"testing"

	"github.com/stretchr/testify/require"
	rtb "go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func Test_ValidateSettings_Scenarios(t *testing.T) {
	cases := []struct {
		name    string
		req     domain.UpdateRunnerConfigReq
		wantErr error
	}{
		{"nothing set is valid", domain.UpdateRunnerConfigReq{}, nil},
		{
			"every supported value is valid",
			domain.UpdateRunnerConfigReq{
				PullPolicy:          rtb.NewOptional([]string{pullPolicyAlways, "if-not-present"}),
				AllowedPullPolicies: rtb.NewOptional([]string{"never"}),
				CheckInterval:       rtb.NewOptional(int32(5)),
				LogLevel:            rtb.NewOptional("debug"),
				ShutdownTimeout:     rtb.NewOptional(int32(30)),
			},
			nil,
		},
		{
			"empty list, zero numbers and empty log level clear a key",
			domain.UpdateRunnerConfigReq{
				PullPolicy:      rtb.NewOptional([]string{}),
				CheckInterval:   rtb.NewOptional(int32(0)),
				LogLevel:        rtb.NewOptional(""),
				ShutdownTimeout: rtb.NewOptional(int32(0)),
			},
			nil,
		},
		{
			"unknown pull policy",
			domain.UpdateRunnerConfigReq{PullPolicy: rtb.NewOptional([]string{"sometimes"})},
			user_errors.ErrRunnerSettingsInvalid,
		},
		{
			"empty pull policy element",
			domain.UpdateRunnerConfigReq{PullPolicy: rtb.NewOptional([]string{""})},
			user_errors.ErrRunnerSettingsInvalid,
		},
		{
			"duplicate pull policy",
			domain.UpdateRunnerConfigReq{PullPolicy: rtb.NewOptional([]string{pullPolicyAlways, pullPolicyAlways})},
			user_errors.ErrRunnerSettingsInvalid,
		},
		{
			"unknown allowed pull policy",
			domain.UpdateRunnerConfigReq{AllowedPullPolicies: rtb.NewOptional([]string{"Always"})},
			user_errors.ErrRunnerSettingsInvalid,
		},
		{
			"unknown log level",
			domain.UpdateRunnerConfigReq{LogLevel: rtb.NewOptional("trace")},
			user_errors.ErrRunnerSettingsInvalid,
		},
		{
			"negative check interval",
			domain.UpdateRunnerConfigReq{CheckInterval: rtb.NewOptional(int32(-1))},
			user_errors.ErrRunnerSettingsInvalid,
		},
		{
			"negative shutdown timeout",
			domain.UpdateRunnerConfigReq{ShutdownTimeout: rtb.NewOptional(int32(-1))},
			user_errors.ErrRunnerSettingsInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSettings(tc.req)
			if tc.wantErr == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func Test_MergeSettings_OnlyValidFieldsOverride(t *testing.T) {
	current := domain.GitlabRunnerSettings{
		PullPolicy:      []string{pullPolicyAlways},
		CheckInterval:   3,
		LogLevel:        "info",
		ShutdownTimeout: 10,
	}
	req := domain.UpdateRunnerConfigReq{
		PullPolicy: rtb.NewOptional([]string{}),
		LogLevel:   rtb.NewOptional("debug"),
	}

	got := mergeSettings(current, req)

	require.Empty(t, got.PullPolicy)
	require.Equal(t, "debug", got.LogLevel)
	require.Equal(t, int32(3), got.CheckInterval)
	require.Equal(t, int32(10), got.ShutdownTimeout)
}
