import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"

import {TaskStatus, TaskStatusStatus} from "@/app/api/velez"
import RegistryDeployProgressScreen from "@/dialogs/RegistryInstanceCreateDialog/screens/RegistryDeployProgressScreen.tsx"
import {WatchTaskStream} from "@/processes/api/tasks.ts"

vi.mock("@/processes/api/tasks.ts", () => ({
    WatchTaskStream: vi.fn(),
}))

interface RenderOptions {
    onSuccess?: () => void
    onClose?: () => void
}

function renderScreen({onSuccess, onClose}: RenderOptions = {}) {
    const start = vi.fn(() => (
        Promise.resolve({entityId: "entity-1", action: "create_registry_instance"})
    ))

    render(
        <RegistryDeployProgressScreen
            name="my-registry"
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

describe("RegistryDeployProgressScreen", () => {
    it("shows the currently running step, then success once the task completes", async () => {
        const onSuccess = vi.fn()

        vi.mocked(WatchTaskStream).mockImplementation(async (_req, onStatus) => {
            onStatus({
                status: TaskStatusStatus.RUNNING,
                jobs: [{name: "deploy_registry", status: TaskStatusStatus.RUNNING}],
            } as TaskStatus)
            onStatus({status: TaskStatusStatus.DONE, jobs: []} as TaskStatus)
        })

        renderScreen({onSuccess})

        await waitFor(() => {
            expect(screen.getByText("Deploy registry…")).toBeInTheDocument()
        })

        await waitFor(() => {
            expect(screen.getByText("Registry deployed")).toBeInTheDocument()
        })
        expect(onSuccess).toHaveBeenCalledTimes(1)
        expect(WatchTaskStream).toHaveBeenCalledWith(
            {entityId: "entity-1", action: "create_registry_instance"},
            expect.any(Function)
        )
    })

    it("shows the error and does not call onSuccess when the task fails", async () => {
        const onSuccess = vi.fn()

        vi.mocked(WatchTaskStream).mockImplementation(async (_req, onStatus) => {
            onStatus({status: TaskStatusStatus.FAILED, error: "disk full", jobs: []} as TaskStatus)
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
            onStatus({status: TaskStatusStatus.FAILED, error: "boom", jobs: []} as TaskStatus)
        })

        renderScreen({onClose})

        const closeButton = await screen.findByText("Close")
        fireEvent.click(closeButton)

        expect(onClose).toHaveBeenCalledTimes(1)
    })
})
