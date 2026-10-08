import {beforeEach, describe, expect, it, vi} from "vitest"

import {TaskStatus, TaskStatusStatus} from "@/app/api/velez"
import {useTaskWatch} from "@/app/hooks/taskWatch/TaskWatch.ts"
import {WatchTaskStream} from "@/processes/api/tasks.ts"

vi.mock("@/processes/api/tasks.ts", () => ({WatchTaskStream: vi.fn()}))

const KEY = "runner/set_runner_buildkit"

function flush() {
    return new Promise((resolve) => setTimeout(resolve, 0))
}

function streamThatEmitsThenRejects(status: TaskStatusStatus) {
    vi.mocked(WatchTaskStream).mockImplementation((_req, onStatus) => {
        onStatus({status} as Partial<TaskStatus> as TaskStatus)
        return Promise.reject(new Error("stream broken"))
    })
}

describe("useTaskWatch", () => {
    beforeEach(() => {
        useTaskWatch.setState({statusByTask: {}, watchingByTask: {}})
        vi.mocked(WatchTaskStream).mockReset()
    })

    it("clears a non-terminal status when the stream rejects", async () => {
        streamThatEmitsThenRejects(TaskStatusStatus.RUNNING)

        useTaskWatch.getState().watch("runner", "set_runner_buildkit")
        await flush()

        expect(useTaskWatch.getState().statusByTask[KEY]).toBeUndefined()
        expect(useTaskWatch.getState().watchingByTask[KEY]).toBe(false)
    })

    it("keeps a terminal status when the stream rejects", async () => {
        streamThatEmitsThenRejects(TaskStatusStatus.DONE)

        useTaskWatch.getState().watch("runner", "set_runner_buildkit")
        await flush()

        expect(useTaskWatch.getState().statusByTask[KEY]?.status).toBe(TaskStatusStatus.DONE)
    })
})
