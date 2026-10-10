import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import CreateDindDialogForm
    from "@/dialogs/CreateServiceDialog/screens/DindScreen/components/CreateDindDialogForm/CreateDindDialogForm.tsx"

function inputFor(label: string) {
    return screen.getByText(label).previousElementSibling as HTMLInputElement
}

function renderForm() {
    const onSubmit = vi.fn()
    const onCancel = vi.fn()

    render(<CreateDindDialogForm onSubmit={onSubmit} onCancel={onCancel}/>)
    return {onSubmit, onCancel}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("CreateDindDialogForm", () => {
    it("disables Create while the name is blank", () => {
        renderForm()

        expect(screen.getByRole("button", {name: "Create"})).toBeDisabled()

        fireEvent.change(inputFor("Name"), {target: {value: "   "}})
        expect(screen.getByRole("button", {name: "Create"})).toBeDisabled()
    })

    it("shows the name error and disables Create when the name is invalid", () => {
        renderForm()

        fireEvent.change(inputFor("Name"), {target: {value: "Bad Name"}})

        expect(screen.getByRole("alert")).toHaveTextContent("Instance name must be 2-32 characters")
        expect(screen.getByRole("button", {name: "Create"})).toBeDisabled()
    })

    it("submits the trimmed name with Sysbox enabled when Create is clicked", () => {
        const {onSubmit} = renderForm()
        fireEvent.change(inputFor("Name"), {target: {value: " ci "}})

        fireEvent.click(screen.getByRole("button", {name: "Create"}))

        expect(onSubmit).toHaveBeenCalledWith({name: "ci", environment: undefined, isSysboxEnabled: true})
    })

    it("submits Sysbox disabled after the user turns the toggle off", () => {
        const {onSubmit} = renderForm()
        fireEvent.change(inputFor("Name"), {target: {value: "ci"}})
        fireEvent.click(screen.getByRole("checkbox", {name: "Sysbox isolation"}))

        fireEvent.click(screen.getByRole("button", {name: "Create"}))

        expect(onSubmit.mock.calls[0][0].isSysboxEnabled).toBe(false)
    })

    it("calls onCancel when Cancel is clicked", () => {
        const {onCancel} = renderForm()

        fireEvent.click(screen.getByRole("button", {name: "Cancel"}))

        expect(onCancel).toHaveBeenCalledTimes(1)
    })
})
