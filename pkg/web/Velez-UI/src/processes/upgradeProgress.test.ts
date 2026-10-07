import {describe, expect, it} from "vitest"

import {TaskStatus, TaskStatusStatus} from "@/app/api/velez"
import {jobLabel, upgradeProgress} from "@/processes/upgradeProgress.ts"

const {PENDING, RUNNING, DONE, FAILED} = TaskStatusStatus

function task(status: TaskStatusStatus, jobs: Array<[string, TaskStatusStatus]> = [], error = ""): TaskStatus {
    return {status, error, jobs: jobs.map(([name, jobStatus]) => ({name, status: jobStatus}))}
}

describe("upgradeProgress", () => {
    it("does report nothing in flight when there is no task", () => {
        expect(upgradeProgress(undefined)).toEqual({isInFlight: false, percent: 0, stepLabel: "", isFailed: false})
    })

    it("does report Queued at zero percent when pending without jobs", () => {
        expect(upgradeProgress(task(PENDING))).toEqual({
            isInFlight: true, percent: 0, stepLabel: "Queued", isFailed: false,
        })
    })

    it("does compute the rounded percent of done jobs", () => {
        const progress = upgradeProgress(task(RUNNING, [
            ["check_self_upgrade", DONE],
            ["capture_old_container", RUNNING],
            ["discover_sidecars", PENDING],
        ]))

        expect(progress.percent).toBe(33)
        expect(progress.isInFlight).toBe(true)
    })

    it("does pick the running job as the current step", () => {
        const progress = upgradeProgress(task(RUNNING, [
            ["check_self_upgrade", DONE],
            ["prepare_image", RUNNING],
            ["healthcheck", PENDING],
        ]))

        expect(progress.stepLabel).toBe("Pulling image")
    })

    it("does pick the first pending job when none is running", () => {
        const progress = upgradeProgress(task(RUNNING, [
            ["check_self_upgrade", DONE],
            ["healthcheck", PENDING],
            ["sync_addresses", PENDING],
        ]))

        expect(progress.stepLabel).toBe("Waiting for healthcheck")
    })

    it("does pick the last job when all of them are done but the task is still running", () => {
        const progress = upgradeProgress(task(RUNNING, [
            ["check_self_upgrade", DONE],
            ["sync_addresses", DONE],
        ]))

        expect(progress.stepLabel).toBe("Syncing addresses")
        expect(progress.percent).toBe(100)
    })

    it("does report Applied at 100 percent when the task is done", () => {
        expect(upgradeProgress(task(DONE, [["check_self_upgrade", DONE]]))).toEqual({
            isInFlight: false, percent: 100, stepLabel: "Applied", isFailed: false,
        })
    })

    it("does report the error and unlock when the task failed", () => {
        const progress = upgradeProgress(task(FAILED, [
            ["check_self_upgrade", DONE],
            ["prepare_image", FAILED],
        ], "pull denied"))

        expect(progress.isInFlight).toBe(false)
        expect(progress.isFailed).toBe(true)
        expect(progress.stepLabel).toBe("Failed: pull denied")
    })

    it("does report not in flight for an unknown status", () => {
        expect(upgradeProgress(task(TaskStatusStatus.UNKNOWN)).isInFlight).toBe(false)
    })
})

describe("jobLabel", () => {
    it("does fall back to sentence case for a job without a mapped label", () => {
        expect(jobLabel("warm_up_cache")).toBe("Warm up cache")
    })
})
