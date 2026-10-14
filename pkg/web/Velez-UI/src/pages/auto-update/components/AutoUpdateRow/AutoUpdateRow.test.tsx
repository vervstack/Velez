import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"

import {AutoUpdateBumpLevel} from "@/app/api/velez/auto_update_api.pb"
import type {ListAutoUpdateServicesResponseService} from "@/app/api/velez/auto_update_api.pb"
import {TaskStatusStatus} from "@/app/api/velez"
import type {TaskStatus} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useWatchedTask} from "@/app/hooks/taskWatch/WatchedTask.ts"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {SetServiceAutoUpdateMutation} from "@/processes/queries/autoUpdate.ts"
import AutoUpdateRow from "@/pages/auto-update/components/AutoUpdateRow/AutoUpdateRow.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
const watch = vi.hoisted(() => vi.fn())

vi.mock("@/app/hooks/taskWatch/TaskWatch.ts", () => ({
    useTaskWatch: (selector: (state: { watch: typeof watch }) => unknown) => selector({watch}),
}))
vi.mock("@/app/hooks/taskWatch/WatchedTask.ts", () => ({useWatchedTask: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/app/queryClient.ts", () => ({queryClient: {invalidateQueries: vi.fn()}}))
vi.mock("@/processes/queries/autoUpdate.ts", () => ({
    SetServiceAutoUpdateMutation: vi.fn(),
    AUTO_UPDATE_QUERY_KEY: ["auto-update"],
}))

type SetMutation = ReturnType<typeof SetServiceAutoUpdateMutation>
type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>

function task(status: TaskStatusStatus): TaskStatus {
    return {status} as Partial<TaskStatus> as TaskStatus
}

function renderRow(service: ListAutoUpdateServicesResponseService, watched?: TaskStatus) {
    const mutateAsync = vi.fn().mockResolvedValue({})
    const OpenDialog = vi.fn()
    const CloseDialog = vi.fn()
    const catchGrpc = vi.fn()
    vi.mocked(SetServiceAutoUpdateMutation).mockReturnValue({mutateAsync} as Partial<SetMutation> as SetMutation)
    vi.mocked(useDialog).mockReturnValue({OpenDialog, CloseDialog} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({bake: vi.fn(), catchGrpc} as Partial<Toaster> as Toaster)
    vi.mocked(useWatchedTask).mockReturnValue(watched)

    render(<AutoUpdateRow service={service}/>)
    return {mutateAsync, OpenDialog, CloseDialog, catchGrpc}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("AutoUpdateRow", () => {
    it("shows the name, the image and the bump level name", () => {
        renderRow({name: "api", image: "reg/api:1.0.2", isEnabled: false, bumpLevel: AutoUpdateBumpLevel.MINOR})

        expect(screen.getByText("api")).toBeInTheDocument()
        expect(screen.getByText("reg/api:1.0.2")).toBeInTheDocument()
        expect(screen.getByText("Major + minor")).toBeInTheDocument()
    })

    it("opens an enable confirmation that mentions the restart when the toggle is switched on", () => {
        const {OpenDialog, mutateAsync} = renderRow({name: "api", isEnabled: false})

        fireEvent.click(screen.getByRole("switch"))
        render(OpenDialog.mock.calls[0][0])

        expect(screen.getByText("Enable auto-update?")).toBeInTheDocument()
        expect(screen.getByText(/api restarts once/)).toBeInTheDocument()
        expect(mutateAsync).not.toHaveBeenCalled()
    })

    it("enables the service with the default PATCH level, then closes and watches the task", async () => {
        const {OpenDialog, mutateAsync, CloseDialog} = renderRow({name: "api", isEnabled: false})
        fireEvent.click(screen.getByRole("switch"))
        render(OpenDialog.mock.calls[0][0])

        fireEvent.click(screen.getByText("Confirm"))

        await waitFor(() => expect(mutateAsync).toHaveBeenCalledWith({
            name: "api",
            isEnabled: true,
            bumpLevel: AutoUpdateBumpLevel.PATCH,
        }))
        await waitFor(() => expect(CloseDialog).toHaveBeenCalled())
        expect(watch).toHaveBeenCalledWith("api", "upgrade_smerd")
    })

    it("asks to disable and sends the current level when an enabled service is switched off", async () => {
        const {OpenDialog, mutateAsync} = renderRow(
            {name: "api", isEnabled: true, bumpLevel: AutoUpdateBumpLevel.MAJOR}
        )
        fireEvent.click(screen.getByRole("switch"))
        render(OpenDialog.mock.calls[0][0])

        expect(screen.getByText("Disable auto-update?")).toBeInTheDocument()
        fireEvent.click(screen.getByText("Confirm"))

        await waitFor(() => expect(mutateAsync).toHaveBeenCalledWith({
            name: "api",
            isEnabled: false,
            bumpLevel: AutoUpdateBumpLevel.MAJOR,
        }))
    })

    it("opens a bump level confirmation when the level of an enabled service changes", async () => {
        const {OpenDialog, mutateAsync} = renderRow(
            {name: "api", isEnabled: true, bumpLevel: AutoUpdateBumpLevel.PATCH}
        )

        fireEvent.click(screen.getByText("All versions (major, minor, patch)"))
        fireEvent.mouseDown(screen.getByText("Major only"))
        render(OpenDialog.mock.calls[0][0])

        expect(screen.getByText("Change bump level?")).toBeInTheDocument()
        fireEvent.click(screen.getByText("Confirm"))

        await waitFor(() => expect(mutateAsync).toHaveBeenCalledWith({
            name: "api",
            isEnabled: true,
            bumpLevel: AutoUpdateBumpLevel.MAJOR,
        }))
    })

    it("changes only the local level without a dialog when the service is disabled", () => {
        const {OpenDialog} = renderRow({name: "api", isEnabled: false, bumpLevel: AutoUpdateBumpLevel.PATCH})

        fireEvent.click(screen.getByText("All versions (major, minor, patch)"))
        fireEvent.mouseDown(screen.getByText("Major only"))

        expect(OpenDialog).not.toHaveBeenCalled()
        expect(screen.getByText("Major only")).toBeInTheDocument()
    })

    it("routes a rejected change through the toaster error surface", async () => {
        const {OpenDialog, mutateAsync, catchGrpc} = renderRow({name: "api", isEnabled: false})
        const failure = new Error("boom")
        mutateAsync.mockRejectedValue(failure)
        fireEvent.click(screen.getByRole("switch"))
        render(OpenDialog.mock.calls[0][0])

        fireEvent.click(screen.getByText("Confirm"))

        await waitFor(() => expect(catchGrpc).toHaveBeenCalledWith(failure))
    })

    it("disables the toggle and shows Restarting while a task is running", () => {
        renderRow({name: "api", isEnabled: true}, task(TaskStatusStatus.RUNNING))

        expect(screen.getByText("Restarting…")).toBeInTheDocument()
        expect(screen.getByRole("switch")).toBeDisabled()
        expect(screen.queryByText("Major only")).not.toBeInTheDocument()
    })

    it("keeps the toggle enabled once the task is done", () => {
        renderRow({name: "api", isEnabled: true}, task(TaskStatusStatus.DONE))

        expect(screen.queryByText("Restarting…")).not.toBeInTheDocument()
        expect(screen.getByRole("switch")).not.toBeDisabled()
    })
})
