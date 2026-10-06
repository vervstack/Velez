import {RunnerLogLevel, RunnerPullPolicy} from "@/app/api/velez"

export interface PolicyOption {
    id: RunnerPullPolicy
    name: string
}

export interface LogLevelOption {
    id: RunnerLogLevel
    name: string
}

export const MAX_PULL_POLICY_CHAIN = 3

export const PULL_POLICY_OPTIONS: PolicyOption[] = [
    {id: RunnerPullPolicy.RUNNER_PULL_POLICY_ALWAYS, name: "Always"},
    {id: RunnerPullPolicy.RUNNER_PULL_POLICY_IF_NOT_PRESENT, name: "If not present"},
    {id: RunnerPullPolicy.RUNNER_PULL_POLICY_NEVER, name: "Never"},
]

export const LOG_LEVEL_OPTIONS: LogLevelOption[] = [
    {id: RunnerLogLevel.RUNNER_LOG_LEVEL_UNSPECIFIED, name: "Default"},
    {id: RunnerLogLevel.RUNNER_LOG_LEVEL_DEBUG, name: "Debug"},
    {id: RunnerLogLevel.RUNNER_LOG_LEVEL_INFO, name: "Info"},
    {id: RunnerLogLevel.RUNNER_LOG_LEVEL_WARN, name: "Warn"},
    {id: RunnerLogLevel.RUNNER_LOG_LEVEL_ERROR, name: "Error"},
    {id: RunnerLogLevel.RUNNER_LOG_LEVEL_FATAL, name: "Fatal"},
    {id: RunnerLogLevel.RUNNER_LOG_LEVEL_PANIC, name: "Panic"},
]

const WHOLE_NUMBER = /^\d+$/

export function parseNonNegativeInt(value: string): number | undefined {
    const trimmed = value.trim()
    if (trimmed === "") return 0
    if (!WHOLE_NUMBER.test(trimmed)) return undefined

    const parsed = Number(trimmed)
    if (!Number.isSafeInteger(parsed)) return undefined

    return parsed
}

export function availablePullPolicies(chain: RunnerPullPolicy[], index: number): PolicyOption[] {
    const chosenElsewhere = chain.filter((_, i) => i !== index)
    return PULL_POLICY_OPTIONS.filter((option) => !chosenElsewhere.includes(option.id))
}

export function nextFallbackPolicy(chain: RunnerPullPolicy[]): RunnerPullPolicy | undefined {
    return PULL_POLICY_OPTIONS.find((option) => !chain.includes(option.id))?.id
}

export function canAddFallback(chain: RunnerPullPolicy[]): boolean {
    return chain.length < MAX_PULL_POLICY_CHAIN && nextFallbackPolicy(chain) !== undefined
}

export function toggleAllowedPolicy(
    current: RunnerPullPolicy[],
    policy: RunnerPullPolicy,
    isChecked: boolean,
): RunnerPullPolicy[] {
    const withoutPolicy = current.filter((p) => p !== policy)
    const next = isChecked ? [...withoutPolicy, policy] : withoutPolicy

    return PULL_POLICY_OPTIONS.map((option) => option.id).filter((id) => next.includes(id))
}

export function arePolicyListsEqual(a: RunnerPullPolicy[], b: RunnerPullPolicy[]): boolean {
    return a.length === b.length && a.every((policy, i) => policy === b[i])
}
