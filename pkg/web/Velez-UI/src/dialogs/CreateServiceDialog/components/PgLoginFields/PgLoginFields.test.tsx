import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import PgLoginFields from "@/dialogs/CreateServiceDialog/components/PgLoginFields/PgLoginFields.tsx"

function renderFields() {
    const onSuperuserChange = vi.fn()
    const onPasswordChange = vi.fn()

    const {container} = render(
        <PgLoginFields
            superuser=""
            password=""
            onSuperuserChange={onSuperuserChange}
            onPasswordChange={onPasswordChange}
        />
    )
    return {container, onSuperuserChange, onPasswordChange}
}

describe("PgLoginFields", () => {
    it("masks the password and reports edits", () => {
        const {container, onPasswordChange} = renderFields()

        const password = container.querySelector("input[type=password]") as HTMLInputElement
        fireEvent.change(password, {target: {value: "secret"}})

        expect(screen.getByText("Password")).toBeInTheDocument()
        expect(onPasswordChange).toHaveBeenCalledWith("secret")
    })

    it("reports the superuser as typed", () => {
        const {onSuperuserChange} = renderFields()

        fireEvent.change(screen.getByRole("textbox"), {target: {value: "admin"}})

        expect(screen.getByText("Superuser")).toBeInTheDocument()
        expect(onSuperuserChange).toHaveBeenCalledWith("admin")
    })
})
