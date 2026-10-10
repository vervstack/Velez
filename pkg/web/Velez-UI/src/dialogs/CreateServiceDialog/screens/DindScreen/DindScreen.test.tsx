import {useState} from "react"
import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {CreateDindMutation} from "@/processes/queries/dinds.ts"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import DindScreen from "@/dialogs/CreateServiceDialog/screens/DindScreen/DindScreen.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/dinds.ts", () => ({
    CreateDindMutation: vi.fn(),
    DINDS_QUERY_KEY: ["dinds"],
}))
vi.mock("@/dialogs/CreateServiceDialog/screens/DindScreen/components/CreateDindDialogForm/CreateDindDialogForm.tsx", () => ({
    default: function FormStub({onSubmit, onCancel}: { onSubmit(req: { name: string }): void, onCancel(): void }) {
        const [name, setName] = useState("")
        return (
            <>
                <Input label="daemon name" inputValue={name} onChange={setName}/>
                <Button variant="primary" onClick={() => onSubmit({name})}>form submit</Button>
                <Button variant="secondary" onClick={onCancel}>form cancel</Button>
            </>
        )
    },
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
type Toaster = ReturnType<typeof useToaster>
type CreateMutation = ReturnType<typeof CreateDindMutation>

function nameInput() {
    return screen.getByText("daemon name").previousElementSibling as HTMLInputElement
}

function fillName(value: string) {
    fireEvent.change(nameInput(), {target: {value}})
}

function renderScreen() {
    const CloseDialog = vi.fn()
    const onBusyChange = vi.fn()
    const mutateAsync = vi.fn().mockResolvedValue({entityId: "ci", action: "create_dind"})
    vi.mocked(useDialog).mockReturnValue({CloseDialog} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({bake: vi.fn(), catchGrpc: vi.fn()} as Partial<Toaster> as Toaster)
    vi.mocked(CreateDindMutation).mockReturnValue(
        {mutateAsync, isPending: false} as Partial<CreateMutation> as CreateMutation
    )

    render(<DindScreen onBusyChange={onBusyChange}/>)
    return {CloseDialog, mutateAsync, onBusyChange}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("DindScreen", () => {
    it("swaps to the progress screen and creates the daemon when the form is submitted", () => {
        const {mutateAsync} = renderScreen()

        fillName("ci")
        fireEvent.click(screen.getByText("form submit"))
        fireEvent.click(screen.getByText("start task"))

        expect(mutateAsync).toHaveBeenCalledWith({name: "ci"})
    })

    it("hides the form while the progress screen is shown", () => {
        renderScreen()
        fillName("ci")

        fireEvent.click(screen.getByText("form submit"))

        expect(screen.getByText("daemon name")).not.toBeVisible()
        expect(screen.getByText("start task")).toBeVisible()
    })

    it("shows the form again with the typed name when Back to form is clicked", () => {
        renderScreen()
        fillName("ci")
        fireEvent.click(screen.getByText("form submit"))

        fireEvent.click(screen.getByText("back to form"))

        expect(screen.getByText("daemon name")).toBeVisible()
        expect(nameInput()).toHaveValue("ci")
        expect(screen.queryByText("start task")).not.toBeInTheDocument()
    })

    it("reports busy while the progress screen is shown and idle again after going back", () => {
        const {onBusyChange} = renderScreen()
        fillName("ci")

        fireEvent.click(screen.getByText("form submit"))

        expect(onBusyChange).toHaveBeenLastCalledWith(true)

        fireEvent.click(screen.getByText("back to form"))

        expect(onBusyChange).toHaveBeenLastCalledWith(false)
    })

    it("closes the dialog when the form is cancelled", () => {
        const {CloseDialog} = renderScreen()

        fireEvent.click(screen.getByText("form cancel"))

        expect(CloseDialog).toHaveBeenCalledTimes(1)
    })
})
