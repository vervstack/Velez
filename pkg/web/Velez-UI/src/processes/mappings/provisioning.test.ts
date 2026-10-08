import {describe, expect, it} from "vitest"

import {ProvisioningTaskStatus} from "@/app/api/velez/velez_common.pb"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {
    currentStepName,
    isInstanceProvisioning,
    isProvisioningActive,
    isProvisioningFailed,
    provisioningRefetchInterval,
    stepProgress,
} from "@/processes/mappings/provisioning.ts"

function newTask(task: Partial<ProvisioningTask>): ProvisioningTask {
    return {entityId: "main", action: "create", status: ProvisioningTaskStatus.RUNNING, ...task}
}

describe("isProvisioningActive", () => {
    it("is true for pending and running tasks", () => {
        expect(isProvisioningActive(newTask({status: ProvisioningTaskStatus.PENDING}))).toBe(true)
        expect(isProvisioningActive(newTask({status: ProvisioningTaskStatus.RUNNING}))).toBe(true)
    })

    it("is false for done, failed and unspecified tasks", () => {
        expect(isProvisioningActive(newTask({status: ProvisioningTaskStatus.DONE}))).toBe(false)
        expect(isProvisioningActive(newTask({status: ProvisioningTaskStatus.FAILED}))).toBe(false)
        expect(isProvisioningActive(newTask({status: undefined}))).toBe(false)
    })
})

describe("isProvisioningFailed", () => {
    it("is true only for failed tasks", () => {
        expect(isProvisioningFailed(newTask({status: ProvisioningTaskStatus.FAILED}))).toBe(true)
        expect(isProvisioningFailed(newTask({status: ProvisioningTaskStatus.RUNNING}))).toBe(false)
        expect(isProvisioningFailed(newTask({status: ProvisioningTaskStatus.DONE}))).toBe(false)
    })
})

describe("provisioningRefetchInterval", () => {
    it("polls every 2 seconds when any task is active", () => {
        const tasks = [newTask({status: ProvisioningTaskStatus.FAILED}), newTask({})]

        expect(provisioningRefetchInterval(tasks)).toBe(2000)
    })

    it("stops polling when no task is active or the list is missing", () => {
        expect(provisioningRefetchInterval([newTask({status: ProvisioningTaskStatus.FAILED})])).toBe(false)
        expect(provisioningRefetchInterval([])).toBe(false)
        expect(provisioningRefetchInterval(undefined)).toBe(false)
    })
})

describe("currentStepName", () => {
    it("returns the first running job", () => {
        const task = newTask({
            jobs: [
                {name: "pull", status: ProvisioningTaskStatus.DONE},
                {name: "start", status: ProvisioningTaskStatus.RUNNING},
                {name: "probe", status: ProvisioningTaskStatus.PENDING},
            ],
        })

        expect(currentStepName(task)).toBe("start")
    })

    it("falls back to the first pending job when none is running", () => {
        const task = newTask({
            jobs: [
                {name: "pull", status: ProvisioningTaskStatus.DONE},
                {name: "start", status: ProvisioningTaskStatus.PENDING},
            ],
        })

        expect(currentStepName(task)).toBe("start")
    })

    it("returns an empty string without jobs or active steps", () => {
        expect(currentStepName(newTask({}))).toBe("")
        expect(currentStepName(newTask({jobs: [{name: "pull", status: ProvisioningTaskStatus.DONE}]}))).toBe("")
    })
})

describe("stepProgress", () => {
    it("counts done jobs against the total", () => {
        const task = newTask({
            jobs: [
                {name: "pull", status: ProvisioningTaskStatus.DONE},
                {name: "start", status: ProvisioningTaskStatus.DONE},
                {name: "probe", status: ProvisioningTaskStatus.RUNNING},
            ],
        })

        expect(stepProgress(task)).toEqual({done: 2, total: 3})
    })

    it("returns zeros without jobs", () => {
        expect(stepProgress(newTask({}))).toEqual({done: 0, total: 0})
    })
})

describe("isInstanceProvisioning", () => {
    const tasks = [newTask({entityId: "main"}), newTask({entityId: "old", status: ProvisioningTaskStatus.FAILED})]

    it("matches an active task by entity id", () => {
        expect(isInstanceProvisioning("main", tasks)).toBe(true)
    })

    it("matches an active task by prefixed entity id", () => {
        expect(isInstanceProvisioning("pg-main", tasks, "pg-")).toBe(true)
    })

    it("ignores inactive tasks and unrelated names", () => {
        expect(isInstanceProvisioning("old", tasks)).toBe(false)
        expect(isInstanceProvisioning("other", tasks)).toBe(false)
        expect(isInstanceProvisioning("pg-main", tasks)).toBe(false)
    })
})
