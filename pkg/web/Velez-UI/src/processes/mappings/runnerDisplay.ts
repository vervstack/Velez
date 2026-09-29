// Mirrors labels.GitlabRunnerNamePrefix in internal/domain/labels/verv_labels.go —
// the single source of truth this string must match.
const GITLAB_RUNNER_NAME_PREFIX = "gitlab_runner_"

export function isGitlabRunnerName(name: string): boolean {
    return name.startsWith(GITLAB_RUNNER_NAME_PREFIX)
}

export function deriveRunnerDisplayName(name: string): string {
    return isGitlabRunnerName(name) ? name.slice(GITLAB_RUNNER_NAME_PREFIX.length) : name
}
