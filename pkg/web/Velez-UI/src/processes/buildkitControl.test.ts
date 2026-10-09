import {describe, expect, it} from "vitest"

import {TaskStatus, TaskStatusStatus} from "@/app/api/velez"
import {buildkitControl, isTaskTerminal} from "@/processes/buildkitControl.ts"

function task(status: TaskStatusStatus): TaskStatus {
    return {status} as Partial<TaskStatus> as TaskStatus
}

describe("buildkitControl", () => {
    it("locks and shows the progress label while the task is pending", () => {
        expect(buildkitControl(task(TaskStatusStatus.PENDING), false)).toEqual({isLocked: true, label: "BuildKit…", hoverLabel: "BuildKit…"})
    })

    it("locks while the task is running", () => {
        expect(buildkitControl(task(TaskStatusStatus.RUNNING), true).isLocked).toBe(true)
    })

    it("is unlocked, reads on and offers to disable when enabled and no task is in flight", () => {
        expect(buildkitControl(undefined, true))
            .toEqual({isLocked: false, label: "BuildKit: on", hoverLabel: "Disable BuildKit"})
    })

    it("is unlocked, reads off and offers to enable after a finished task", () => {
        expect(buildkitControl(task(TaskStatusStatus.DONE), false))
            .toEqual({isLocked: false, label: "BuildKit: off", hoverLabel: "Enable BuildKit"})
    })

    it("unlocks after a failed task", () => {
        expect(buildkitControl(task(TaskStatusStatus.FAILED), true).isLocked).toBe(false)
    })
})

describe("isTaskTerminal", () => {
    it("is true for done and failed only", () => {
        expect(isTaskTerminal(task(TaskStatusStatus.DONE))).toBe(true)
        expect(isTaskTerminal(task(TaskStatusStatus.FAILED))).toBe(true)
        expect(isTaskTerminal(task(TaskStatusStatus.RUNNING))).toBe(false)
        expect(isTaskTerminal(undefined)).toBe(false)
    })
})
