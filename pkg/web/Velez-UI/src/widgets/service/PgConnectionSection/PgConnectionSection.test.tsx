import {describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"

import PgConnectionSection from "@/widgets/service/PgConnectionSection/PgConnectionSection.tsx"
import {useListPgInstancesQuery} from "@/processes/queries/pg_instances.ts"

vi.mock("@/processes/queries/pg_instances.ts", () => ({useListPgInstancesQuery: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({
    useToaster: () => ({catchGrpc: vi.fn(), bake: vi.fn()}),
}))
vi.mock("@/widgets/PgInstanceCredentials/PgInstanceCredentials.tsx", () => ({
    default: ({name}: {name: string}) => <div>credentials of {name}</div>,
}))

function renderSection(serviceName: string) {
    vi.mocked(useListPgInstancesQuery).mockReturnValue(
        {
            data: {instances: [{name: "pgaas_orders", dbName: "orders", username: "app", port: 5432, exposedPort: 18714}]},
        } as Partial<ReturnType<typeof useListPgInstancesQuery>> as ReturnType<typeof useListPgInstancesQuery>,
    )

    return render(<PgConnectionSection serviceName={serviceName}/>)
}

describe("PgConnectionSection", () => {
    it("renders the connection section with credentials when the service is a postgres instance", () => {
        renderSection("pgaas_orders")

        expect(screen.getByText("Connection")).toBeInTheDocument()
        expect(screen.getByText("5432 → 18714")).toBeInTheDocument()
        expect(screen.getByText("credentials of pgaas_orders")).toBeInTheDocument()
    })

    it("renders nothing when no instance matches the service name", () => {
        const {container} = renderSection("other_service")

        expect(container).toBeEmptyDOMElement()
    })
})
