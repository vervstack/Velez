import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import AasHintBanner from "@/components/AasHintBanner/AasHintBanner.tsx"

function renderBanner() {
    const onOpen = vi.fn()
    const onDismiss = vi.fn()
    render(<AasHintBanner title="Postgres" onOpen={onOpen} onDismiss={onDismiss}/>)
    return {onOpen, onDismiss}
}

describe("AasHintBanner", () => {
    it("shows the managed option title when rendered", () => {
        renderBanner()

        expect(screen.getByRole("note")).toHaveTextContent("There is a managed option for this image: Postgres")
    })

    it("calls onOpen when Open is clicked", () => {
        const {onOpen, onDismiss} = renderBanner()

        fireEvent.click(screen.getByRole("button", {name: "Open"}))

        expect(onOpen).toHaveBeenCalledTimes(1)
        expect(onDismiss).not.toHaveBeenCalled()
    })

    it("calls onDismiss when Dismiss is clicked", () => {
        const {onOpen, onDismiss} = renderBanner()

        fireEvent.click(screen.getByRole("button", {name: "Dismiss"}))

        expect(onDismiss).toHaveBeenCalledTimes(1)
        expect(onOpen).not.toHaveBeenCalled()
    })
})
