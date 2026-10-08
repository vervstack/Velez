import {afterEach, describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"

import {ProvisioningTaskStatus} from "@/app/api/velez/velez_common.pb"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useListPgInstancesQuery} from "@/processes/queries/pg_instances.ts"
import PostgresPage from "@/pages/postgres/PostgresPage.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/pg_instances.ts", () => ({
    useListPgInstancesQuery: vi.fn(),
    PG_INSTANCES_QUERY_KEY: ["pg-instances"],
}))
vi.mock("@/widgets/ProvisioningRow/ProvisioningRow.tsx", () => ({
    default: ({task, title}: { task: ProvisioningTask, title: string }) => <span>{title}: {task.entityId}</span>,
}))
vi.mock("@/pages/postgres/components/PgInstanceRow/PgInstanceRow.tsx", () => ({
    default: ({instance}: { instance: { name?: string } }) => <span>instance {instance.name}</span>,
}))
vi.mock("@/dialogs/CreateServiceDialog/CreateServiceDialog.tsx", () => ({default: () => null}))

type PgQuery = ReturnType<typeof useListPgInstancesQuery>
type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>

function renderPage(data: { instances?: { name: string }[], provisioning?: ProvisioningTask[] }) {
    vi.mocked(useListPgInstancesQuery).mockReturnValue({data} as Partial<PgQuery> as PgQuery)
    vi.mocked(useDialog).mockReturnValue({OpenDialog: vi.fn()} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({catchGrpc: vi.fn()} as Partial<Toaster> as Toaster)

    render(<PostgresPage/>)
}

const RUNNING_TASK: ProvisioningTask = {taskId: "1", entityId: "orders", status: ProvisioningTaskStatus.RUNNING}

afterEach(() => {
    vi.clearAllMocks()
})

describe("PostgresPage", () => {
    it("shows a provisioning row and no empty state when only a create task exists", () => {
        renderPage({instances: [], provisioning: [RUNNING_TASK]})

        expect(screen.getByText("Creating database: orders")).toBeInTheDocument()
        expect(screen.queryByText("No Postgres instances on this node.")).not.toBeInTheDocument()
    })

    it("shows the empty state when there are no instances and no tasks", () => {
        renderPage({instances: [], provisioning: []})

        expect(screen.getByText("No Postgres instances on this node.")).toBeInTheDocument()
    })

    it("hides the prefixed instance row of a database that is still being created", () => {
        renderPage({instances: [{name: "pgaas_orders"}, {name: "pgaas_billing"}], provisioning: [RUNNING_TASK]})

        expect(screen.queryByText("instance pgaas_orders")).not.toBeInTheDocument()
        expect(screen.getByText("instance pgaas_billing")).toBeInTheDocument()
        expect(screen.getByText("2 instances")).toBeInTheDocument()
    })
})
