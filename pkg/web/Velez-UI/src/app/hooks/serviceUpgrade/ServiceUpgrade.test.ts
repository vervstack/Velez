import {beforeEach, describe, expect, it, vi} from "vitest"

import {TaskStatus, TaskStatusStatus} from "@/app/api/velez"
import {useServiceUpgrade} from "@/app/hooks/serviceUpgrade/ServiceUpgrade.ts"
import {WatchServiceUpgradeStream} from "@/processes/api/tasks.ts"

vi.mock("@/processes/api/tasks.ts", () => ({WatchServiceUpgradeStream: vi.fn()}))

function flush() {
    return new Promise((resolve) => setTimeout(resolve, 0))
}

function streamThatEmitsThenRejects(status: TaskStatusStatus) {
    vi.mocked(WatchServiceUpgradeStream).mockImplementation((_name, onStatus) => {
        onStatus({status} as Partial<TaskStatus> as TaskStatus)
        return Promise.reject(new Error("stream broken"))
    })
}

describe("useServiceUpgrade", () => {
    beforeEach(() => {
        useServiceUpgrade.setState({statusByService: {}, watchingByService: {}})
        vi.mocked(WatchServiceUpgradeStream).mockReset()
    })

    it("clears a non-terminal status when the stream rejects", async () => {
        streamThatEmitsThenRejects(TaskStatusStatus.RUNNING)

        useServiceUpgrade.getState().watch("svc")
        await flush()

        expect(useServiceUpgrade.getState().statusByService["svc"]).toBeUndefined()
        expect(useServiceUpgrade.getState().watchingByService["svc"]).toBe(false)
    })

    it("keeps a terminal status when the stream rejects", async () => {
        streamThatEmitsThenRejects(TaskStatusStatus.FAILED)

        useServiceUpgrade.getState().watch("svc")
        await flush()

        expect(useServiceUpgrade.getState().statusByService["svc"]?.status).toBe(TaskStatusStatus.FAILED)
    })
})
