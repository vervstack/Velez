package runners_api_impl

import (
	"testing"

	"github.com/stretchr/testify/require"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func Test_PullPolicy_RoundTrip(t *testing.T) {
	policies := []pb.RunnerPullPolicy{
		pb.RunnerPullPolicy_RUNNER_PULL_POLICY_ALWAYS,
		pb.RunnerPullPolicy_RUNNER_PULL_POLICY_IF_NOT_PRESENT,
		pb.RunnerPullPolicy_RUNNER_PULL_POLICY_NEVER,
	}

	got := pullPolicyEnums(pullPolicyStrings(policies))

	require.Equal(t, policies, got)
}

func Test_PullPolicyStrings_UnspecifiedMapsToEmpty(t *testing.T) {
	got := pullPolicyStrings([]pb.RunnerPullPolicy{pb.RunnerPullPolicy_RUNNER_PULL_POLICY_UNSPECIFIED})

	require.Equal(t, []string{""}, got)
}

func Test_PullPolicyStrings_EmptyListIsNotNil(t *testing.T) {
	require.NotNil(t, pullPolicyStrings(nil))
}

func Test_PullPolicyEnums_UnknownIsDropped(t *testing.T) {
	got := pullPolicyEnums([]string{"always", "sometimes", ""})

	require.Equal(t, []pb.RunnerPullPolicy{pb.RunnerPullPolicy_RUNNER_PULL_POLICY_ALWAYS}, got)
}

func Test_LogLevel_RoundTrip(t *testing.T) {
	levels := []pb.RunnerLogLevel{
		pb.RunnerLogLevel_RUNNER_LOG_LEVEL_DEBUG,
		pb.RunnerLogLevel_RUNNER_LOG_LEVEL_INFO,
		pb.RunnerLogLevel_RUNNER_LOG_LEVEL_WARN,
		pb.RunnerLogLevel_RUNNER_LOG_LEVEL_ERROR,
		pb.RunnerLogLevel_RUNNER_LOG_LEVEL_FATAL,
		pb.RunnerLogLevel_RUNNER_LOG_LEVEL_PANIC,
	}

	for _, level := range levels {
		require.Equal(t, level, logLevelEnum(logLevelString(level)))
	}
}

func Test_LogLevel_UnspecifiedAndUnknown(t *testing.T) {
	require.Equal(t, "", logLevelString(pb.RunnerLogLevel_RUNNER_LOG_LEVEL_UNSPECIFIED))
	require.Equal(t, pb.RunnerLogLevel_RUNNER_LOG_LEVEL_UNSPECIFIED, logLevelEnum("trace"))
}
