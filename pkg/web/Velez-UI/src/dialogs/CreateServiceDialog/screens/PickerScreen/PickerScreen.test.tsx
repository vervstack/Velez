import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import PickerScreen from "@/dialogs/CreateServiceDialog/screens/PickerScreen/PickerScreen.tsx"
import {ServiceScreen} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"

function renderPicker(suggestedScreen?: ServiceScreen) {
    const onSelect = vi.fn()
    render(<PickerScreen suggestedScreen={suggestedScreen} onSelect={onSelect}/>)
    return {onSelect}
}

describe("PickerScreen", () => {
    it("renders a card for every product", () => {
        renderPicker()

        expect(screen.getByText("Generic container/image")).toBeInTheDocument()
        expect(screen.getByText("PostgreSQL")).toBeInTheDocument()
        expect(screen.getByText("Container registry")).toBeInTheDocument()
        expect(screen.getByText("GitHub runner")).toBeInTheDocument()
        expect(screen.getByText("GitLab runner")).toBeInTheDocument()
    })

    it("calls onSelect with the matching screen when a card is clicked", () => {
        const {onSelect} = renderPicker()

        fireEvent.click(screen.getByText("PostgreSQL"))
        fireEvent.click(screen.getByText("GitLab runner"))

        expect(onSelect).toHaveBeenNthCalledWith(1, "postgres")
        expect(onSelect).toHaveBeenNthCalledWith(2, "gitlabRunner")
    })

    it("shows the Suggested badge only when a screen is suggested", () => {
        renderPicker("registry")

        expect(screen.getAllByText("Suggested")).toHaveLength(1)
    })

    it("shows no badge without a suggestion", () => {
        renderPicker()

        expect(screen.queryByText("Suggested")).not.toBeInTheDocument()
    })
})
