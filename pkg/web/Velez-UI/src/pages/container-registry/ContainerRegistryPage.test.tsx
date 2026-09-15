import {afterEach, describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"
import {MemoryRouter} from "react-router-dom"

import ContainerRegistryPage from "@/pages/container-registry/ContainerRegistryPage.tsx"
import {useListRegistryInstancesQuery} from "@/processes/queries/registry_instances.ts"
import type {RegistryInstance} from "@/app/api/velez"

vi.mock("@/processes/queries/registry_instances.ts", () => ({
    useListRegistryInstancesQuery: vi.fn(),
}))

function mockInstances(instances: RegistryInstance[]) {
    vi.mocked(useListRegistryInstancesQuery).mockReturnValue({
        data: {instances},
        isLoading: false,
        error: null,
    } as Partial<ReturnType<typeof useListRegistryInstancesQuery>> as ReturnType<typeof useListRegistryInstancesQuery>)
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("ContainerRegistryPage", () => {
    it("shows the empty state when there are no registry instances", () => {
        mockInstances([])

        render(<MemoryRouter><ContainerRegistryPage/></MemoryRouter>)

        expect(screen.getByText("No container registries on this node.")).toBeInTheDocument()
    })

    it("renders one row per registry instance, sorted by name", () => {
        mockInstances([
            {name: "zeta", environment: "prod", port: 5000, username: "verv", status: "running"},
            {name: "alpha", environment: "dev", port: 5001, username: "verv", status: "running"},
        ])

        render(<MemoryRouter><ContainerRegistryPage/></MemoryRouter>)

        const names = screen.getAllByText(/^(alpha|zeta)$/).map((el) => el.textContent)
        expect(names).toEqual(["alpha", "zeta"])
        expect(screen.getByText("2 instances")).toBeInTheDocument()
    })
})
