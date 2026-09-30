import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import RegistryLoginFields
    from "@/dialogs/CreateServiceDialog/components/RegistryLoginFields/RegistryLoginFields.tsx"

function renderFields() {
    const onUsernameChange = vi.fn()
    const onPasswordChange = vi.fn()

    const {container} = render(
        <RegistryLoginFields
            username=""
            password=""
            onUsernameChange={onUsernameChange}
            onPasswordChange={onPasswordChange}
        />
    )
    return {container, onUsernameChange, onPasswordChange}
}

describe("RegistryLoginFields", () => {
    it("masks the password and reports edits", () => {
        const {container, onPasswordChange} = renderFields()

        const password = container.querySelector("input[type=password]") as HTMLInputElement
        fireEvent.change(password, {target: {value: "secret"}})

        expect(screen.getByText("Password")).toBeInTheDocument()
        expect(onPasswordChange).toHaveBeenCalledWith("secret")
    })

    it("reports the username as typed and shows the hint", () => {
        const {onUsernameChange} = renderFields()

        fireEvent.change(screen.getByRole("textbox"), {target: {value: "ci"}})

        expect(screen.getByText("Username")).toBeInTheDocument()
        expect(screen.getByText(/tested before anything is changed/)).toBeInTheDocument()
        expect(onUsernameChange).toHaveBeenCalledWith("ci")
    })
})
