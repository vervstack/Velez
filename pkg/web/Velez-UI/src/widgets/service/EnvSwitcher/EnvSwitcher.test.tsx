import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {SmerdStatus} from "@/app/api/velez"
import type {ServiceEnvironment, ServiceSidecarView} from "@/model/service_page/ServicePageModel"
import {useListServiceEnvsQuery} from "@/processes/queries/services.ts"
import EnvSwitcher from "@/widgets/service/EnvSwitcher/EnvSwitcher.tsx"

const navigate = vi.fn()

vi.mock("react-router-dom", () => ({useNavigate: () => navigate}))
vi.mock("@/processes/queries/services.ts", () => ({useListServiceEnvsQuery: vi.fn()}))

function newEnv(overrides: Partial<ServiceEnvironment>): ServiceEnvironment {
    return {
        id: "prod",
        label: "prod",
        status: "running",
        version: "v1",
        deployedAgo: "1h ago",
        containerId: "",
        containerName: "",
        health: "healthy",
        ...overrides,
    }
}

function renderSwitcher(envs: ServiceEnvironment[], sidecars?: ServiceSidecarView[]) {
    vi.mocked(useListServiceEnvsQuery).mockReturnValue(
        {data: envs} as Partial<ReturnType<typeof useListServiceEnvsQuery>> as ReturnType<typeof useListServiceEnvsQuery>
    )

    render(<EnvSwitcher serviceName="svc" sidecars={sidecars}/>)
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("EnvSwitcher", () => {
    it("navigates to the container page when a card with a container id is clicked", () => {
        renderSwitcher([newEnv({containerId: "abc123"})])

        fireEvent.click(screen.getByText("prod"))

        expect(navigate).toHaveBeenCalledWith("/container/abc123")
    })

    it("does not navigate when the card has no container id", () => {
        renderSwitcher([newEnv({containerId: ""})])

        fireEvent.click(screen.getByText("prod"))

        expect(navigate).not.toHaveBeenCalled()
    })

    it("renders a card with the name, image and sidecar label for each sidecar", () => {
        renderSwitcher([newEnv({})], [
            {containerId: "side1", name: "garage-webui", imageName: "khairul169/garage-webui:1.0", status: SmerdStatus.running},
        ])

        expect(screen.getByText("garage-webui")).toBeTruthy()
        expect(screen.getByText("khairul169/garage-webui:1.0")).toBeTruthy()
        expect(screen.getByText("sidecar")).toBeTruthy()
    })

    it("navigates to the sidecar container page when its card is clicked", () => {
        renderSwitcher([newEnv({})], [
            {containerId: "side1", name: "garage-webui", imageName: "img", status: SmerdStatus.running},
        ])

        fireEvent.click(screen.getByText("garage-webui"))

        expect(navigate).toHaveBeenCalledWith("/container/side1")
    })

    it("hides the empty state when only sidecars exist", () => {
        renderSwitcher([], [
            {containerId: "side1", name: "garage-webui", imageName: "img", status: SmerdStatus.running},
        ])

        expect(screen.queryByText("No environment data available")).toBeNull()
        expect(screen.getByText("garage-webui")).toBeTruthy()
    })
})
