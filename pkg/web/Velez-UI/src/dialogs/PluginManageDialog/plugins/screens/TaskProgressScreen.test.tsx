import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"

import {EnablePluginResponse, TaskStatus, TaskStatusStatus} from "@/app/api/velez"
import TaskProgressScreen from "@/dialogs/PluginManageDialog/plugins/screens/TaskProgressScreen.tsx"
import {WatchTaskStream} from "@/processes/api/tasks.ts"

vi.mock("@/processes/api/tasks.ts", () => ({
    WatchTaskStream: vi.fn(),
}))

const RUNNING_JOBS = [
    {name: "generate_credentials", status: TaskStatusStatus.DONE},
    {name: "create_container", status: TaskStatusStatus.RUNNING},
]

interface RenderOptions {
    onSuccess?: () => void
    onClose?: () => void
}

function renderScreen({onSuccess, onClose}: RenderOptions = {}) {
    const start = vi.fn((): Promise<EnablePluginResponse> => (
        Promise.resolve({entityId: "entity-1", action: "enable_registry"})
    ))

    render(
        <TaskProgressScreen
            title="Enabling registry"
            metaLine="Container registry"
            start={start}
            onSuccess={onSuccess}
            onClose={onClose ?? vi.fn()}
        />
    )

    return {start}
}

afterEach(() => {
    vi.restoreAllMocks()
})

describe("TaskProgressScreen", () => {
    it("shows the task id and step chips once running, then success once the task completes", async () => {
        const onSuccess = vi.fn()

        vi.mocked(WatchTaskStream).mockImplementation(async (_req, onStatus) => {
            onStatus({status: TaskStatusStatus.RUNNING, taskId: "task-5", jobs: RUNNING_JOBS} as TaskStatus)
            onStatus({
                status: TaskStatusStatus.DONE,
                taskId: "task-5",
                jobs: RUNNING_JOBS.map((job) => ({...job, status: TaskStatusStatus.DONE})),
            } as TaskStatus)
        })

        renderScreen({onSuccess})

        await waitFor(() => {
            expect(screen.getByText("Enabling registry · Task #task-5")).toBeInTheDocument()
        })
        expect(screen.getByText("Creating container")).toBeInTheDocument()

        await waitFor(() => {
            expect(screen.getByText("Done")).toBeInTheDocument()
        })
        expect(onSuccess).toHaveBeenCalledTimes(1)
        expect(WatchTaskStream).toHaveBeenCalledWith(
            {entityId: "entity-1", action: "enable_registry"},
            expect.any(Function)
        )
    })

    it("shows the error and does not call onSuccess when the task fails", async () => {
        const onSuccess = vi.fn()

        vi.mocked(WatchTaskStream).mockImplementation(async (_req, onStatus) => {
            onStatus({status: TaskStatusStatus.FAILED, error: "disk full", taskId: "task-9", jobs: []} as TaskStatus)
        })

        renderScreen({onSuccess})

        await waitFor(() => {
            expect(screen.getByText("disk full")).toBeInTheDocument()
        })
        expect(onSuccess).not.toHaveBeenCalled()
    })

    it("calls onClose when Close is clicked after failure", async () => {
        const onClose = vi.fn()

        vi.mocked(WatchTaskStream).mockImplementation(async (_req, onStatus) => {
            onStatus({status: TaskStatusStatus.FAILED, error: "boom", taskId: "task-2", jobs: []} as TaskStatus)
        })

        renderScreen({onClose})

        const closeButton = await screen.findByText("Close")
        fireEvent.click(closeButton)

        expect(onClose).toHaveBeenCalledTimes(1)
    })
})
