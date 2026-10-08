import {afterEach, describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"

import type {S3Instance} from "@/app/api/velez/s3_api.pb"
import {ProvisioningTaskStatus} from "@/app/api/velez/velez_common.pb"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useListS3InstancesQuery} from "@/processes/queries/s3.ts"
import S3Page from "@/pages/s3/S3Page.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/processes/queries/s3.ts", () => ({
    useListS3InstancesQuery: vi.fn(),
    S3_INSTANCES_QUERY_KEY: ["s3-instances"],
}))
vi.mock("@/dialogs/CreateS3InstanceDialog/CreateS3InstanceDialog.tsx", () => ({default: () => null}))
vi.mock("@/pages/s3/components/S3InstanceDetail/S3InstanceDetail.tsx", () => ({default: () => null}))
vi.mock("@/pages/s3/components/S3InstanceRow/S3InstanceRow.tsx", () => ({
    default: ({instance}: { instance: S3Instance }) => <span>instance {instance.name}</span>,
}))
vi.mock("@/pages/s3/components/S3ListSkeleton/S3ListSkeleton.tsx", () => ({
    default: () => <span>s3 skeleton</span>,
}))
vi.mock("@/widgets/ProvisioningRow/ProvisioningRow.tsx", () => ({
    default: ({task}: { task: ProvisioningTask }) => <span>provisioning {task.entityId}</span>,
}))

type S3Query = ReturnType<typeof useListS3InstancesQuery>
type Dialog = ReturnType<typeof useDialog>

function runningCreateTask(entityId: string): ProvisioningTask {
    return {taskId: "1", entityId, action: "create_s3_instance", status: ProvisioningTaskStatus.RUNNING}
}

function renderPage(query: Partial<S3Query>) {
    vi.mocked(useListS3InstancesQuery).mockReturnValue({refetch: vi.fn(), ...query} as Partial<S3Query> as S3Query)
    vi.mocked(useDialog).mockReturnValue({OpenDialog: vi.fn()} as Partial<Dialog> as Dialog)

    render(<S3Page/>)
}

function loaded(instances: S3Instance[], provisioning: ProvisioningTask[] = []): Partial<S3Query> {
    return {data: {instances, provisioning}}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("S3Page", () => {
    it("shows the empty state when there are no instances and no create tasks", () => {
        renderPage(loaded([]))

        expect(screen.getByText("No S3 instances on this node.")).toBeInTheDocument()
    })

    it("renders a provisioning row and no empty state when only a create task exists", () => {
        renderPage(loaded([], [runningCreateTask("main")]))

        expect(screen.getByText("provisioning main")).toBeInTheDocument()
        expect(screen.queryByText("No S3 instances on this node.")).not.toBeInTheDocument()
    })

    it("hides the real row of an instance that is still provisioning but keeps counting it", () => {
        renderPage(loaded([{name: "main"}, {name: "other"}], [runningCreateTask("main")]))

        expect(screen.queryByText("instance main")).not.toBeInTheDocument()
        expect(screen.getByText("instance other")).toBeInTheDocument()
        expect(screen.getByText("2 instances")).toBeInTheDocument()
    })

    it("shows the skeleton while loading", () => {
        renderPage({isLoading: true})

        expect(screen.getByText("s3 skeleton")).toBeInTheDocument()
    })

    it("shows a count skeleton instead of 0 instances while loading", () => {
        renderPage({isLoading: true})

        expect(screen.queryByText("0 instances")).not.toBeInTheDocument()
        expect(screen.getByText("instances")).toHaveAttribute("aria-busy", "true")
    })
})
