import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import PickerScreen from "@/dialogs/CreateServiceDialog/screens/PickerScreen/PickerScreen.tsx"
import {ProductScreen, ServiceScreen} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"

function renderPicker(suggestedScreen?: ServiceScreen, enabledScreens?: ProductScreen[]) {
    const onSelect = vi.fn()
    render(<PickerScreen suggestedScreen={suggestedScreen} enabledScreens={enabledScreens} onSelect={onSelect}/>)
    return {onSelect}
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

    it("disables every card outside enabledScreens with a note", () => {
        const {onSelect} = renderPicker(undefined, ["generic", "postgres"])

        expect(screen.getAllByText("Not available for existing containers yet")).toHaveLength(5)

        fireEvent.click(screen.getByText("Container registry"))
        expect(onSelect).not.toHaveBeenCalled()

        fireEvent.click(screen.getByText("PostgreSQL"))
        expect(onSelect).toHaveBeenCalledWith("postgres")
    })

    it("enables every card and shows no note when enabledScreens is omitted", () => {
        renderPicker()

        expect(screen.queryByText("Not available for existing containers yet")).not.toBeInTheDocument()
    })
})
