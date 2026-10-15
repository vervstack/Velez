import {describe, it, expect, vi, beforeEach, afterEach} from "vitest"
import {render, screen, fireEvent} from "@testing-library/react"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import type {ServiceResource} from "@/model/service_page/ServicePageModel"

import ResourceCard from "./ResourceCard"

const navigate = vi.fn()

vi.mock("react-router-dom", () => ({useNavigate: () => navigate}))
vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/dialogs/ResourceAddressesDialog/ResourceAddressesDialog.tsx", () => ({
    default: () => null,
}))

type Dialog = ReturnType<typeof useDialog>

function renderCard(overrides: Partial<ServiceResource> = {}) {
    const resource: ServiceResource = {
        name: "test-resource",
        type: "postgres",
        status: "healthy",
        icon: "Pg",
        color: "var(--info-color)",
        reconciliation: "unknown",
        addresses: [],
        isOnboarded: false,
        ...overrides,
    }
    const OpenDialog = vi.fn()
    vi.mocked(useDialog).mockReturnValue({OpenDialog} as Partial<Dialog> as Dialog)
    render(<ResourceCard resource={resource}/>)
    return {OpenDialog}
}

describe("ResourceCard", () => {
    let windowOpenSpy: ReturnType<typeof vi.spyOn>

    beforeEach(() => {
        navigate.mockClear()
        windowOpenSpy = vi.spyOn(window, "open").mockReturnValue(null)
    })

    afterEach(() => {
        windowOpenSpy.mockRestore()
    })

    it("navigates to the linked service page when clicked", () => {
        renderCard({linkedServiceName: "dind-main", containerId: "abc"})

        const link = screen.getByRole("link")
        fireEvent.click(link)

        expect(link).toHaveAttribute("title", "Open test-resource")
        expect(navigate).toHaveBeenCalledWith("/service/dind-main")
    })

    it("navigates to the container page when only a container id is set", () => {
        renderCard({containerId: "abc"})

        fireEvent.click(screen.getByRole("link"))

        expect(navigate).toHaveBeenCalledWith("/container/abc")
    })

    it("navigates when Enter is pressed on the card", () => {
        renderCard({linkedServiceName: "dind-main"})

        fireEvent.keyDown(screen.getByRole("link"), {key: "Enter"})

        expect(navigate).toHaveBeenCalledWith("/service/dind-main")
    })

    it("opens a new tab instead of navigating when clicked with the meta key", () => {
        renderCard({linkedServiceName: "dind-main"})

        fireEvent.click(screen.getByRole("link"), {metaKey: true})

        expect(windowOpenSpy).toHaveBeenCalledWith("/service/dind-main", "_blank", "noopener,noreferrer")
        expect(navigate).not.toHaveBeenCalled()
    })

    it("opens a new tab when clicked with the middle button", () => {
        renderCard({containerId: "abc"})

        fireEvent(screen.getByRole("link"), new MouseEvent("auxclick", {bubbles: true, button: 1}))

        expect(windowOpenSpy).toHaveBeenCalledWith("/container/abc", "_blank", "noopener,noreferrer")
        expect(navigate).not.toHaveBeenCalled()
    })

    it("shows the onboarded indicator when the resource is onboarded", () => {
        renderCard({isOnboarded: true})

        expect(screen.getByRole("img", {name: "Onboarded in Velez"})).toBeInTheDocument()
    })

    it("hides the onboarded indicator when the resource is not onboarded", () => {
        renderCard({isOnboarded: false})

        expect(screen.queryByRole("img", {name: "Onboarded in Velez"})).not.toBeInTheDocument()
    })

    it("opens the addresses dialog without navigating when Addresses is clicked", () => {
        const {OpenDialog} = renderCard({
            linkedServiceName: "dind-main",
            addresses: [{host: "h", port: 1, scope: "vcn"}],
        })

        fireEvent.click(screen.getByRole("button", {name: "Addresses"}))

        expect(OpenDialog).toHaveBeenCalledTimes(1)
        expect(navigate).not.toHaveBeenCalled()
    })

    it("opens the web UI without navigating when Web UI is clicked", () => {
        renderCard({linkedServiceName: "dind-main", webUiPort: 3909})

        fireEvent.click(screen.getByRole("button", {name: "Web UI"}))

        expect(windowOpenSpy).toHaveBeenCalledWith(
            expect.stringContaining(":3909"),
            "_blank",
            "noopener,noreferrer",
        )
        expect(navigate).not.toHaveBeenCalled()
    })

    it("does not navigate when Enter is pressed on an action button", () => {
        renderCard({linkedServiceName: "dind-main", webUiPort: 3909})

        fireEvent.keyDown(screen.getByRole("button", {name: "Web UI"}), {key: "Enter"})

        expect(navigate).not.toHaveBeenCalled()
    })

    it("is not a link when there is neither a service nor a container to open", () => {
        renderCard({webUiPort: 3909})

        expect(screen.queryByRole("link")).not.toBeInTheDocument()
    })

    it("shows no actions when there are no addresses and no web UI", () => {
        renderCard()

        expect(screen.queryByRole("button")).not.toBeInTheDocument()
    })
})
