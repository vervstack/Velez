import type {GetSysboxStatusResponse, RunSysboxSmokeTestResponse} from "@/app/api/velez/settings_api.pb"

import type {SysboxStepId} from "@/widgets/settings/SandboxSettings/components/SysboxSetupChecklist/processes/sysboxSetupSteps.ts"

export type SysboxCheckState = "pass" | "fail" | "unknown"

export interface SysboxCheck {
    id: SysboxStepId
    state: SysboxCheckState
    hint?: string
}

const MIN_KERNEL_MAJOR = 5
const MIN_KERNEL_MINOR = 12

function fromBoolean(isPassed: boolean): SysboxCheckState {
    return isPassed ? "pass" : "fail"
}

function deriveKernel(kernelVersion: string | undefined): SysboxCheck {
    const match = /^(\d+)\.(\d+)/.exec(kernelVersion ?? "")
    if (!match) {
        return {id: "kernel", state: "unknown"}
    }

    const major = Number(match[1])
    const minor = Number(match[2])
    const isRecent = major > MIN_KERNEL_MAJOR || (major === MIN_KERNEL_MAJOR && minor >= MIN_KERNEL_MINOR)
    if (isRecent) {
        return {id: "kernel", state: "pass"}
    }
    return {id: "kernel", state: "fail", hint: "older kernels need shiftfs"}
}

function deriveDocker(status: GetSysboxStatusResponse): SysboxCheck {
    if (status.isRootless) {
        return {id: "docker", state: "fail", hint: "Docker is running rootless"}
    }
    if (status.isSnap) {
        return {id: "docker", state: "fail", hint: "Docker is installed as a snap"}
    }
    return {id: "docker", state: "pass"}
}

function deriveEnable(isSysboxEnabled: boolean | undefined, status: GetSysboxStatusResponse | undefined): SysboxCheck {
    if (isSysboxEnabled === undefined) {
        return {id: "enable", state: "unknown"}
    }
    if (!isSysboxEnabled || !status) {
        return {id: "enable", state: fromBoolean(isSysboxEnabled)}
    }

    const onSysbox = status.containersOnSysbox ?? 0
    const total = status.containersTotal ?? 0
    return {
        id: "enable",
        state: "pass",
        hint: `${onSysbox} of ${total} containers run under Sysbox — recreate the rest to move them`,
    }
}

export function deriveSysboxChecks(
    status: GetSysboxStatusResponse | undefined,
    smokeTest: RunSysboxSmokeTestResponse | undefined,
    isSysboxEnabled: boolean | undefined,
): SysboxCheck[] {
    const smokeState: SysboxCheckState = smokeTest ? fromBoolean(smokeTest.isPassed ?? false) : "unknown"
    const smokeHint = smokeTest && !smokeTest.isPassed ? smokeTest.failure : undefined

    const runtimeState: SysboxCheckState = status ? fromBoolean(status.isRuntimeRegistered ?? false) : "unknown"

    return [
        {id: "linux", state: status ? fromBoolean(status.osType === "linux") : "unknown"},
        status ? deriveKernel(status.kernelVersion) : {id: "kernel", state: "unknown"},
        status ? deriveDocker(status) : {id: "docker", state: "unknown"},
        {id: "install", state: runtimeState},
        {id: "services", state: smokeState},
        {id: "runtime", state: runtimeState},
        {id: "smoke", state: smokeState, hint: smokeHint},
        deriveEnable(isSysboxEnabled, status),
        {id: "dind", state: "unknown"},
    ]
}

export function summarizeSysboxChecks(checks: SysboxCheck[]): {passed: number, total: number} {
    const countable = checks.filter(check => check.id !== "dind")
    const passed = countable.filter(check => check.state === "pass").length
    return {passed, total: countable.length}
}
