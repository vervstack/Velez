import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import type {ServiceEnvironment} from "@/model/service_page/ServicePageModel"
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

function renderSwitcher(envs: ServiceEnvironment[]) {
    vi.mocked(useListServiceEnvsQuery).mockReturnValue(
        {data: envs} as Partial<ReturnType<typeof useListServiceEnvsQuery>> as ReturnType<typeof useListServiceEnvsQuery>
    )

    render(<EnvSwitcher serviceName="svc"/>)
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
})
