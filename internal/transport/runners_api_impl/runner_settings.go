package runners_api_impl

import (
	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

var (
	pullPolicyByEnum = map[pb.RunnerPullPolicy]string{
		pb.RunnerPullPolicy_RUNNER_PULL_POLICY_ALWAYS:         "always",
		pb.RunnerPullPolicy_RUNNER_PULL_POLICY_IF_NOT_PRESENT: "if-not-present",
		pb.RunnerPullPolicy_RUNNER_PULL_POLICY_NEVER:          "never",
	}

	logLevelByEnum = map[pb.RunnerLogLevel]string{
		pb.RunnerLogLevel_RUNNER_LOG_LEVEL_DEBUG: "debug",
		pb.RunnerLogLevel_RUNNER_LOG_LEVEL_INFO:  "info",
		pb.RunnerLogLevel_RUNNER_LOG_LEVEL_WARN:  "warn",
		pb.RunnerLogLevel_RUNNER_LOG_LEVEL_ERROR: "error",
		pb.RunnerLogLevel_RUNNER_LOG_LEVEL_FATAL: "fatal",
		pb.RunnerLogLevel_RUNNER_LOG_LEVEL_PANIC: "panic",
	}
)

// pullPolicyStrings maps an unspecified or unknown enum to "", which the
// service rejects as an invalid policy. The result is never nil, so a set but
// empty list stays distinguishable from an unset one.
func pullPolicyStrings(values []pb.RunnerPullPolicy) []string {
	result := make([]string, 0, len(values))

	for _, value := range values {
		result = append(result, pullPolicyByEnum[value])
	}

	return result
}

// pullPolicyEnums drops any string that is not a known policy.
func pullPolicyEnums(values []string) []pb.RunnerPullPolicy {
	var result []pb.RunnerPullPolicy

	for _, value := range values {
		for enum, name := range pullPolicyByEnum {
			if name == value {
				result = append(result, enum)
			}
		}
	}

	return result
}

func logLevelString(level pb.RunnerLogLevel) string {
	return logLevelByEnum[level]
}

func logLevelEnum(level string) pb.RunnerLogLevel {
	for enum, name := range logLevelByEnum {
		if name == level {
			return enum
		}
	}

	return pb.RunnerLogLevel_RUNNER_LOG_LEVEL_UNSPECIFIED
}
