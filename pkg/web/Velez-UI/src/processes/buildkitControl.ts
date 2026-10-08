import {TaskStatus, TaskStatusStatus} from "@/app/api/velez"

export const SET_RUNNER_BUILDKIT_ACTION = "set_runner_buildkit"

export const BUILDKIT_CREATE_TOOLTIP = "Runs a BuildKit daemon inside this runner's DinD on a private network. "
    + "CI jobs reach it at tcp://buildkit:1234 (docker buildx --driver remote) "
    + "and keep their build cache between pipelines."

export const BUILDKIT_UNSUPPORTED_TOOLTIP = "Only available for runners whose jobs run in a Velez DinD."

export interface BuildkitControl {
    isLocked: boolean
    label: string
}

export function isTaskInFlight(task: TaskStatus | undefined): boolean {
    return task?.status === TaskStatusStatus.PENDING || task?.status === TaskStatusStatus.RUNNING
}

export function isTaskTerminal(task: TaskStatus | undefined): boolean {
    return task?.status === TaskStatusStatus.DONE || task?.status === TaskStatusStatus.FAILED
}

export function buildkitControl(task: TaskStatus | undefined, isEnabled: boolean): BuildkitControl {
    if (isTaskInFlight(task)) {
        return {isLocked: true, label: "BuildKit…"}
    }

    return {isLocked: false, label: isEnabled ? "BuildKit: on" : "BuildKit: off"}
}
