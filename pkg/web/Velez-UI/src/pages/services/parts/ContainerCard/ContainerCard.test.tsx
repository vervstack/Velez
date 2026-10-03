import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {ServicePattern} from "@/app/api/velez"
import type {DockerContainer} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import ContainerCard from "@/pages/services/parts/ContainerCard/ContainerCard.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/dialogs/CreateServiceDialog/CreateServiceDialog.tsx", () => ({default: () => null}))
vi.mock("@/processes/queries/containers.ts", () => ({
    CONTAINERS_QUERY_KEY: ["containers"],
    CONTAINER_QUERY_KEY: ["container"],
    FinishOnboardingMutation: () => ({mutateAsync: vi.fn()}),
}))

function renderCard(container: DockerContainer) {
    const OpenDialog = vi.fn()
    vi.mocked(useDialog).mockReturnValue(
        {OpenDialog} as Partial<ReturnType<typeof useDialog>> as ReturnType<typeof useDialog>
    )
    const onOpen = vi.fn()

    render(<ContainerCard container={container} onOpen={onOpen} onFilterByService={vi.fn()}/>)
    return {OpenDialog, onOpen}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("ContainerCard", () => {
    it("offers Register for a container that is not linked to a service", () => {
        renderCard({id: "c1", name: "db"})

        expect(screen.getByRole("button", {name: "Onboard"})).toBeInTheDocument()
    })

    it("offers no Register for a container linked to a service", () => {
        renderCard({id: "c1", name: "db", linkedServiceName: "db-service"})

        expect(screen.queryByRole("button", {name: "Onboard"})).not.toBeInTheDocument()
    })

    it("offers no Register for a registered container that has no linked service", () => {
        renderCard({id: "c1", name: "velez", isRegistered: true})

        expect(screen.queryByRole("button", {name: "Onboard"})).not.toBeInTheDocument()
    })

    it("offers no Register when the register action is hidden", () => {
        const OpenDialog = vi.fn()
        vi.mocked(useDialog).mockReturnValue(
            {OpenDialog} as Partial<ReturnType<typeof useDialog>> as ReturnType<typeof useDialog>
        )

        render(
            <ContainerCard
                container={{id: "c1", name: "db"}}
                onOpen={vi.fn()}
                onFilterByService={vi.fn()}
                isRegisterHidden
            />
        )

        expect(screen.queryByRole("button", {name: "Onboard"})).not.toBeInTheDocument()
    })

    it("opens the adopt dialog for the container without navigating to it", () => {
        const container = {id: "c1", name: "db"}
        const {OpenDialog, onOpen} = renderCard(container)

        fireEvent.click(screen.getByRole("button", {name: "Onboard"}))

        expect(OpenDialog).toHaveBeenCalledTimes(1)
        expect(OpenDialog.mock.calls[0][0].props.container).toBe(container)
        expect(onOpen).not.toHaveBeenCalled()
    })

    it("hints at the suggested pattern of an unlinked container", () => {
        renderCard({id: "c1", name: "db", suggestedPattern: ServicePattern.SERVICE_PATTERN_POSTGRES})

        expect(screen.getByText("Looks like PostgreSQL")).toBeInTheDocument()
    })

    it("offers Finish onboarding instead of Register for a leftover container", () => {
        renderCard({id: "c1", name: "db_old", replacedByContainerId: "c2"})

        expect(screen.getByText("waiting")).toBeInTheDocument()
        expect(screen.getByRole("button", {name: "Finish onboarding"})).toBeInTheDocument()
        expect(screen.queryByRole("button", {name: "Onboard"})).not.toBeInTheDocument()
    })

    it("asks for confirmation when Finish onboarding is clicked", () => {
        const {OpenDialog, onOpen} = renderCard({id: "c1", name: "db_old", replacedByContainerId: "c2"})

        fireEvent.click(screen.getByRole("button", {name: "Finish onboarding"}))

        expect(OpenDialog).toHaveBeenCalledTimes(1)
        expect(OpenDialog.mock.calls[0][0].props.title).toBe("Finish onboarding?")
        expect(onOpen).not.toHaveBeenCalled()
    })
})
