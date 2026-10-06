import {describe, it, expect, vi, beforeEach, afterEach} from "vitest"
import {render, screen, fireEvent} from "@testing-library/react"

import type {ServiceResource} from "@/model/service_page/ServicePageModel"

import ResourceCard from "./ResourceCard"

function renderCard(overrides: Partial<ServiceResource> = {}) {
    const resource: ServiceResource = {
        name: "test-resource",
        type: "postgres",
        status: "healthy",
        icon: "Pg",
        color: "var(--info-color)",
        reconciliation: "unknown",
        ...overrides,
    }
    render(<ResourceCard resource={resource}/>)
}

describe("ResourceCard", () => {
    let windowOpenSpy: ReturnType<typeof vi.spyOn>

    beforeEach(() => {
        windowOpenSpy = vi.spyOn(window, "open").mockReturnValue(null)
    })

    afterEach(() => {
        windowOpenSpy.mockRestore()
    })

    it("opens web UI when clicked with webUiPort", () => {
        renderCard({webUiPort: 3909})

        const link = screen.getByRole("link")
        fireEvent.click(link)

        expect(windowOpenSpy).toHaveBeenCalledWith(
            expect.stringContaining(":3909"),
            "_blank",
            "noopener,noreferrer"
        )
    })

    it("does not have role link when webUiPort is undefined", () => {
        renderCard({webUiPort: undefined})

        expect(screen.queryByRole("link")).not.toBeInTheDocument()
    })

    it("does not call window.open when clicked without webUiPort", () => {
        renderCard({webUiPort: undefined})

        const container = screen.getByText("test-resource").closest("div")!.parentElement
        if (container) {
            fireEvent.click(container)
        }

        expect(windowOpenSpy).not.toHaveBeenCalled()
    })

    it("uses webUiHost when it is set", () => {
        renderCard({webUiPort: 3909, webUiHost: "example.com"})

        const link = screen.getByRole("link")
        fireEvent.click(link)

        expect(windowOpenSpy).toHaveBeenCalledWith(
            expect.stringContaining("example.com"),
            "_blank",
            "noopener,noreferrer"
        )
    })

    it("opens web UI when Enter key is pressed", () => {
        renderCard({webUiPort: 3909})

        const link = screen.getByRole("link")
        fireEvent.keyDown(link, {key: "Enter"})

        expect(windowOpenSpy).toHaveBeenCalled()
    })

    it("does not have title when webUiPort is undefined", () => {
        renderCard({webUiPort: undefined})

        const container = screen.getByText("test-resource").closest("div")!.parentElement
        expect(container).not.toHaveAttribute("title")
    })

    it("has title when webUiPort is set", () => {
        renderCard({webUiPort: 3909})

        const link = screen.getByRole("link")
        expect(link).toHaveAttribute("title", "Open web UI")
    })
})
