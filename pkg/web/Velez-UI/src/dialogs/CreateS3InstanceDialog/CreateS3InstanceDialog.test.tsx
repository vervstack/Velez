import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {ListEnvironmentsQuery} from "@/processes/queries/control_plane.ts"
import {CreateS3InstanceMutation} from "@/processes/queries/s3.ts"
import Button from "@/components/base/Button.tsx"
import CreateS3InstanceDialog from "@/dialogs/CreateS3InstanceDialog/CreateS3InstanceDialog.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/processes/queries/control_plane.ts", () => ({ListEnvironmentsQuery: vi.fn()}))
vi.mock("@/processes/queries/s3.ts", () => ({
    CreateS3InstanceMutation: vi.fn(),
    S3_INSTANCES_QUERY_KEY: ["s3-instances"],
}))
vi.mock("@/widgets/TaskProgressScreen/TaskProgressScreen.tsx", () => ({
    default: ({start, onBack}: { start(): Promise<unknown>, onBack(): void }) => (
        <>
            <Button onClick={() => void start()}>start task</Button>
            <Button onClick={onBack}>back to form</Button>
        </>
    ),
}))

type Dialog = ReturnType<typeof useDialog>
type Environments = ReturnType<typeof ListEnvironmentsQuery>
type CreateMutation = ReturnType<typeof CreateS3InstanceMutation>

function inputFor(label: string) {
    return screen.getByText(label).previousElementSibling as HTMLInputElement
}

function renderDialog() {
    const CloseDialog = vi.fn()
    const mutateAsync = vi.fn().mockResolvedValue({entityId: "main", action: "create_s3_instance"})
    vi.mocked(useDialog).mockReturnValue({CloseDialog} as Partial<Dialog> as Dialog)
    vi.mocked(ListEnvironmentsQuery).mockReturnValue(
        {data: {environments: []}} as Partial<Environments> as Environments
    )
    vi.mocked(CreateS3InstanceMutation).mockReturnValue(
        {mutateAsync, isPending: false} as Partial<CreateMutation> as CreateMutation
    )

    render(<CreateS3InstanceDialog/>)
    return {mutateAsync, CloseDialog}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("CreateS3InstanceDialog", () => {
    it("disables Create until a name is entered", () => {
        renderDialog()

        expect(screen.getByRole("button", {name: "Create"})).toBeDisabled()

        fireEvent.change(inputFor("Name"), {target: {value: "main"}})

        expect(screen.getByRole("button", {name: "Create"})).toBeEnabled()
    })

    it("shows the replication factor locked to 1 with a multi-node hint", () => {
        renderDialog()

        const replication = inputFor("Replication factor")

        expect(replication).toHaveValue("1")
        expect(replication).toBeDisabled()
        expect(screen.getByText(/multi-node Garage clusters come later/)).toBeInTheDocument()
    })

    it("keeps the web UI sidecar off by default", () => {
        renderDialog()

        expect(screen.getByRole("checkbox", {name: "Enable web UI"})).not.toBeChecked()
    })

    it("swaps to the progress screen and creates the instance with the web UI flag when Create is clicked", () => {
        const {mutateAsync} = renderDialog()

        fireEvent.change(inputFor("Name"), {target: {value: "main"}})
        fireEvent.click(screen.getByRole("checkbox", {name: "Enable web UI"}))
        fireEvent.click(screen.getByRole("button", {name: "Create"}))
        fireEvent.click(screen.getByRole("button", {name: "start task"}))

        expect(mutateAsync).toHaveBeenCalledWith(expect.objectContaining({
            name: "main",
            replicationFactor: 1,
            enableWebUi: true,
        }))
    })

    it("shows the name error and keeps Create disabled when the name is invalid", () => {
        renderDialog()

        fireEvent.change(inputFor("Name"), {target: {value: "Main Bucket"}})

        expect(screen.getByRole("alert")).toHaveTextContent("Instance name must be 2-32 characters")
        expect(screen.getByRole("button", {name: "Create"})).toBeDisabled()
    })

    it("shows the form again with the entered fields when the progress screen goes back", () => {
        renderDialog()

        fireEvent.change(inputFor("Name"), {target: {value: "main"}})
        fireEvent.click(screen.getByRole("button", {name: "Create"}))

        expect(screen.queryByRole("button", {name: "Create"})).not.toBeInTheDocument()

        fireEvent.click(screen.getByRole("button", {name: "back to form"}))

        expect(inputFor("Name")).toHaveValue("main")
        expect(screen.getByRole("button", {name: "Create"})).toBeEnabled()
        expect(screen.queryByRole("button", {name: "start task"})).not.toBeInTheDocument()
    })

    it("closes the dialog when Cancel is clicked", () => {
        const {CloseDialog} = renderDialog()

        fireEvent.click(screen.getByRole("button", {name: "Cancel"}))

        expect(CloseDialog).toHaveBeenCalledTimes(1)
    })
})
