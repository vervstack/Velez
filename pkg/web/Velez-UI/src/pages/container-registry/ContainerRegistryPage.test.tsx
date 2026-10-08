import {afterEach, describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"
import {MemoryRouter} from "react-router-dom"

import ContainerRegistryPage from "@/pages/container-registry/ContainerRegistryPage.tsx"
import {useListRegistryInstancesQuery} from "@/processes/queries/registry_instances.ts"
import type {RegistryInstance} from "@/app/api/velez"
import {ProvisioningTaskStatus} from "@/app/api/velez/velez_common.pb"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"

vi.mock("@/processes/queries/registry_instances.ts", () => ({
    useListRegistryInstancesQuery: vi.fn(),
    REGISTRY_INSTANCES_QUERY_KEY: ["registry-instances"],
}))
vi.mock("@/widgets/ProvisioningRow/ProvisioningRow.tsx", () => ({
    default: ({task}: { task: ProvisioningTask }) => <span>provisioning {task.entityId}</span>,
}))

function runningCreateTask(entityId: string): ProvisioningTask {
    return {taskId: "1", entityId, action: "create_registry_instance", status: ProvisioningTaskStatus.RUNNING}
}

function mockInstances(instances: RegistryInstance[], provisioning: ProvisioningTask[] = []) {
    vi.mocked(useListRegistryInstancesQuery).mockReturnValue({
        data: {instances, provisioning},
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

    it("renders a provisioning row and no empty state when only a create task exists", () => {
        mockInstances([], [runningCreateTask("main")])

        render(<MemoryRouter><ContainerRegistryPage/></MemoryRouter>)

        expect(screen.getByText("provisioning main")).toBeInTheDocument()
        expect(screen.queryByText("No container registries on this node.")).not.toBeInTheDocument()
    })

    it("hides the real row of a registry whose cr_-prefixed service is still provisioning", () => {
        mockInstances(
            [{name: "cr_main", environment: "dev", port: 5000, username: "verv", status: "running"}],
            [runningCreateTask("main")],
        )

        render(<MemoryRouter><ContainerRegistryPage/></MemoryRouter>)

        expect(screen.queryByText("cr_main")).not.toBeInTheDocument()
        expect(screen.getByText("1 instances")).toBeInTheDocument()
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

    it("shows a count skeleton instead of 0 instances while loading", () => {
        vi.mocked(useListRegistryInstancesQuery).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null,
        } as Partial<ReturnType<typeof useListRegistryInstancesQuery>> as
            ReturnType<typeof useListRegistryInstancesQuery>)

        render(<MemoryRouter><ContainerRegistryPage/></MemoryRouter>)

        expect(screen.queryByText("0 instances")).not.toBeInTheDocument()
        expect(screen.getByText("instances")).toHaveAttribute("aria-busy", "true")
    })
})
