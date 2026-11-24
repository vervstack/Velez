import {describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"

import type {GetPgReplicationStatusResponse} from "@/app/api/velez"
import {useListPgInstancesQuery, usePgReplicationStatusQuery} from "@/processes/queries/pg_instances.ts"
import PgConnectionSection from "@/widgets/service/PgConnectionSection/PgConnectionSection.tsx"

vi.mock("@/processes/queries/pg_instances.ts", () => ({
    useListPgInstancesQuery: vi.fn(),
    usePgReplicationStatusQuery: vi.fn(),
}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({
    useToaster: () => ({catchGrpc: vi.fn(), bake: vi.fn()}),
}))
vi.mock("@/widgets/PgInstanceCredentials/PgInstanceCredentials.tsx", () => ({
    default: ({name}: {name: string}) => <div>credentials of {name}</div>,
}))

function renderSection(serviceName: string, status: GetPgReplicationStatusResponse = {}) {
    vi.mocked(useListPgInstancesQuery).mockReturnValue(
        {
            data: {
                instances: [{name: "pgaas_orders", dbName: "orders", username: "app", port: 5432, exposedPort: 18714}],
            },
        } as Partial<ReturnType<typeof useListPgInstancesQuery>> as ReturnType<typeof useListPgInstancesQuery>,
    )
    const statusResult: Partial<ReturnType<typeof usePgReplicationStatusQuery>> = {data: status}
    vi.mocked(usePgReplicationStatusQuery).mockReturnValue(
        statusResult as ReturnType<typeof usePgReplicationStatusQuery>,
    )

    return render(<PgConnectionSection serviceName={serviceName}/>)
}

describe("PgConnectionSection", () => {
    it("renders the connection section with credentials when the service is a postgres instance", () => {
        renderSection("pgaas_orders")

        expect(screen.getByText("Connection")).toBeInTheDocument()
        expect(screen.getByText("5432 → 18714")).toBeInTheDocument()
        expect(screen.getByText("credentials of pgaas_orders")).toBeInTheDocument()
        expect(screen.getByRole("button", {name: "Set up replication"})).toBeInTheDocument()
    })

    it("shows the upstream and sync badge when replication is confirmed", () => {
        renderSection("pgaas_orders", {isReplica: true, inRecovery: true, upstream: "pgaas_master", inSync: true})

        expect(screen.getByText("Replica of pgaas_master")).toBeInTheDocument()
        expect(screen.getByText("In sync")).toBeInTheDocument()
    })

    it("shows each primary replica's stream state and lag", () => {
        renderSection("pgaas_orders", {
            replicas: [{applicationName: "pgaas_replica", state: "catchup", lagBytes: "4096", replayLagSeconds: 2.5}],
        })

        expect(screen.getByText("pgaas_replica")).toBeInTheDocument()
        expect(screen.getByText("catchup")).toBeInTheDocument()
        expect(screen.getByText("Lag 4096 B · 2.5 s")).toBeInTheDocument()
        expect(screen.queryByText("In sync")).not.toBeInTheDocument()
    })

    it("renders nothing when no instance matches the service name", () => {
        const {container} = renderSection("other_service")

        expect(container).toBeEmptyDOMElement()
    })
})
