import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"

import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {CreateDindMutation} from "@/processes/queries/dinds.ts"
import CreateDindForm from "@/widgets/CreateDindForm/CreateDindForm.tsx"
import {buildCreateDindRequest} from "@/widgets/CreateDindForm/processes/buildCreateDindRequest.ts"

vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/dinds.ts", () => ({CreateDindMutation: vi.fn()}))

type CreateMutation = ReturnType<typeof CreateDindMutation>
type Toaster = ReturnType<typeof useToaster>

function inputFor(label: string) {
    return screen.getByText(label).previousElementSibling as HTMLInputElement
}

function renderForm(mutation: Partial<CreateMutation> = {}) {
    const mutateAsync = vi.fn().mockResolvedValue({})
    const catchGrpc = vi.fn()
    const onCreated = vi.fn()
    const onCancel = vi.fn()
    vi.mocked(CreateDindMutation).mockReturnValue(
        {mutateAsync, ...mutation} as Partial<CreateMutation> as CreateMutation,
    )
    vi.mocked(useToaster).mockReturnValue({bake: vi.fn(), catchGrpc} as Partial<Toaster> as Toaster)

    render(<CreateDindForm onCreated={onCreated} onCancel={onCancel}/>)
    return {mutateAsync, catchGrpc, onCreated, onCancel}
}

function typeName(name: string) {
    fireEvent.change(inputFor("Name"), {target: {value: name}})
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("CreateDindForm", () => {
    it("defaults the Sysbox toggle to on and explains why it is recommended", () => {
        renderForm()

        expect(screen.getByRole("switch", {name: "Sysbox isolation"})).toBeChecked()
        expect(screen.getByText(/Recommended for public repositories/)).toBeInTheDocument()
    })

    it("disables Create while the name is empty or blank", () => {
        renderForm()

        expect(screen.getByRole("button", {name: "Create"})).toBeDisabled()

        typeName("   ")
        expect(screen.getByRole("button", {name: "Create"})).toBeDisabled()

        typeName("ci")
        expect(screen.getByRole("button", {name: "Create"})).toBeEnabled()
    })

    it("creates the daemon with Sysbox enabled when the toggle is left alone", async () => {
        const {mutateAsync, onCreated} = renderForm()
        typeName(" ci ")

        fireEvent.click(screen.getByRole("button", {name: "Create"}))

        const expected = buildCreateDindRequest({name: "ci", environment: "", isSysboxEnabled: true})
        await waitFor(() => expect(mutateAsync).toHaveBeenCalledWith(expected))
        expect(mutateAsync.mock.calls[0][0].isSysboxEnabled).toBe(true)
        await waitFor(() => expect(onCreated).toHaveBeenCalledWith("ci"))
    })

    it("creates the daemon with Sysbox disabled only after the user turns the toggle off", async () => {
        const {mutateAsync} = renderForm()
        typeName("ci")
        fireEvent.click(screen.getByRole("switch", {name: "Sysbox isolation"}))

        fireEvent.click(screen.getByRole("button", {name: "Create"}))

        const expected = buildCreateDindRequest({name: "ci", environment: "", isSysboxEnabled: false})
        await waitFor(() => expect(mutateAsync).toHaveBeenCalledWith(expected))
        expect(mutateAsync.mock.calls[0][0].isSysboxEnabled).toBe(false)
    })

    it("routes a failed create through the toaster and does not report it as created", async () => {
        const failure = new Error("create failed")
        const {catchGrpc, onCreated} = renderForm({mutateAsync: vi.fn().mockRejectedValue(failure)})
        typeName("ci")

        fireEvent.click(screen.getByRole("button", {name: "Create"}))

        await waitFor(() => expect(catchGrpc).toHaveBeenCalledWith(failure))
        expect(onCreated).not.toHaveBeenCalled()
    })

    it("calls onCancel when Cancel is clicked", () => {
        const {onCancel} = renderForm()

        fireEvent.click(screen.getByRole("button", {name: "Cancel"}))

        expect(onCancel).toHaveBeenCalledTimes(1)
    })
})
