import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {isInstanceProvisioning} from "@/processes/mappings/provisioning.ts"

// Mirrors labels.GitlabRunnerNamePrefix in internal/domain/labels/verv_labels.go —
// the single source of truth this string must match.
const GITLAB_RUNNER_NAME_PREFIX = "gitlab_runner_"
const GITHUB_RUNNER_NAME_PREFIX = "github_runner_"

export function isGitlabRunnerName(name: string): boolean {
    return name.startsWith(GITLAB_RUNNER_NAME_PREFIX)
}

export function deriveRunnerDisplayName(name: string): string {
    return isGitlabRunnerName(name) ? name.slice(GITLAB_RUNNER_NAME_PREFIX.length) : name
}

// The create task's entity id is the bare name the user typed; the runner's name carries a provider prefix.
export function isRunnerProvisioning(runnerName: string, tasks: ProvisioningTask[]): boolean {
    return isInstanceProvisioning(runnerName, tasks, GITHUB_RUNNER_NAME_PREFIX)
        || isInstanceProvisioning(runnerName, tasks, GITLAB_RUNNER_NAME_PREFIX)
}
