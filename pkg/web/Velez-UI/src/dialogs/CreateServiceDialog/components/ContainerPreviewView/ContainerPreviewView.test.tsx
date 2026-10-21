import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {ValueOrigin, type ContainerPreview} from "@/app/api/velez"
import ContainerPreviewView, {ContainerPreviewState}
    from "@/dialogs/CreateServiceDialog/components/ContainerPreviewView/ContainerPreviewView.tsx"

const PREVIEW: ContainerPreview = {
    containerId: "c1",
    networkMode: {key: "mode", value: "container:root", origin: ValueOrigin.ADDED},
    ports: [{key: "80/tcp", value: "8080", origin: ValueOrigin.TRANSFERRED}],
    env: [
        {key: "PLAIN", value: "kept", origin: ValueOrigin.TRANSFERRED},
        {key: "VERV_SERVICE", value: "web", origin: ValueOrigin.ADDED},
    ],
}

function renderView(overrides: Partial<ContainerPreviewState> = {}) {
    const onRetry = vi.fn()
    const state: ContainerPreviewState = {isLoading: false, isError: false, isDisabled: false, onRetry, ...overrides}
    render(<ContainerPreviewView {...state}/>)
    return {onRetry}
}

describe("ContainerPreviewView", () => {
    it("lists the non-empty sections as key = value rows", () => {
        renderView({preview: PREVIEW})

        expect(screen.getByText("Network mode")).toBeInTheDocument()
        expect(screen.getByText("Ports")).toBeInTheDocument()
        expect(screen.getByText("Env")).toBeInTheDocument()
        expect(screen.queryByText("Labels")).not.toBeInTheDocument()
        expect(screen.getByText("PLAIN")).toBeInTheDocument()
        expect(screen.getByText("kept")).toBeInTheDocument()
    })

    it("marks only the rows added by onboarding", () => {
        renderView({preview: PREVIEW})

        expect(screen.getAllByText("added")).toHaveLength(2)
        expect(screen.getByText("VERV_SERVICE").closest("li")).toHaveTextContent("added")
        expect(screen.getByText("PLAIN").closest("li")).not.toHaveTextContent("added")
    })

    it("shows skeleton rows instead of content while loading", () => {
        renderView({isLoading: true})

        expect(screen.queryByText("Env")).not.toBeInTheDocument()
        expect(screen.queryByText("Failed to load the preview.")).not.toBeInTheDocument()
    })

    it("shows an error with a retry that refetches", () => {
        const {onRetry} = renderView({isError: true})

        expect(screen.getByText("Failed to load the preview.")).toBeInTheDocument()
        fireEvent.click(screen.getByRole("button", {name: "Retry"}))

        expect(onRetry).toHaveBeenCalledTimes(1)
    })

    it("renders nothing while the preview is disabled", () => {
        renderView({isDisabled: true, preview: PREVIEW})

        expect(screen.queryByText("Env")).not.toBeInTheDocument()
    })

    it("says there is nothing to show when the container has no entry", () => {
        renderView({preview: undefined})

        expect(screen.getByText("Nothing to show for this container.")).toBeInTheDocument()
    })
})
