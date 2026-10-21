import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import PickerScreen from "@/dialogs/CreateServiceDialog/screens/PickerScreen/PickerScreen.tsx"
import {ProductScreen, ServiceScreen} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"

const NOTE_SELECTOR = "[data-tooltip-content=\"Not available for existing containers yet\"]"

function renderPicker(suggestedScreen?: ServiceScreen, enabledScreens?: ProductScreen[]) {
    const onSelect = vi.fn()
    const {container} = render(
        <PickerScreen suggestedScreen={suggestedScreen} enabledScreens={enabledScreens} onSelect={onSelect}/>,
    )
    return {onSelect, container}
}

describe("PickerScreen", () => {
    it("renders a card for every product", () => {
        renderPicker()

        expect(screen.getByText("Generic container/image")).toBeInTheDocument()
        expect(screen.getByText("PostgreSQL")).toBeInTheDocument()
        expect(screen.getByText("Container registry")).toBeInTheDocument()
        expect(screen.getByText("S3 storage")).toBeInTheDocument()
        expect(screen.getByText("Docker in Docker")).toBeInTheDocument()
        expect(screen.getByText("GitHub runner")).toBeInTheDocument()
        expect(screen.getByText("GitLab runner")).toBeInTheDocument()
    })

    it("calls onSelect with the s3 and dind screens for the new products", () => {
        const {onSelect} = renderPicker()

        fireEvent.click(screen.getByText("S3 storage"))
        fireEvent.click(screen.getByText("Docker in Docker"))

        expect(onSelect).toHaveBeenNthCalledWith(1, "s3")
        expect(onSelect).toHaveBeenNthCalledWith(2, "dind")
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

    it("disables every card outside enabledScreens with a tooltip note", () => {
        const {onSelect, container} = renderPicker(undefined, ["generic", "postgres"])

        expect(container.querySelectorAll(NOTE_SELECTOR)).toHaveLength(5)

        fireEvent.click(screen.getByText("Container registry"))
        expect(onSelect).not.toHaveBeenCalled()

        fireEvent.click(screen.getByText("PostgreSQL"))
        expect(onSelect).toHaveBeenCalledWith("postgres")
    })

    it("marks the selected card active and still reports a click on another card", () => {
        const onSelect = vi.fn()
        render(<PickerScreen selectedScreen="postgres" onSelect={onSelect}/>)

        expect(screen.getAllByText("✓")).toHaveLength(1)

        fireEvent.click(screen.getByText("S3 storage"))
        expect(onSelect).toHaveBeenCalledWith("s3")
    })

    it("enables every card and shows no note when enabledScreens is omitted", () => {
        const {container} = renderPicker()

        expect(container.querySelectorAll(NOTE_SELECTOR)).toHaveLength(0)
    })
})
