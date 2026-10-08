import {afterEach, describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"

import {ProvisioningTaskStatus} from "@/app/api/velez/velez_common.pb"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useListRunnersQuery} from "@/processes/queries/runners.ts"
import RunnersPage from "@/pages/runners/RunnersPage.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/runners.ts", () => ({
    useListRunnersQuery: vi.fn(),
    RUNNERS_QUERY_KEY: ["runners"],
}))
vi.mock("@/widgets/ProvisioningRow/ProvisioningRow.tsx", () => ({
    default: ({task, noun}: { task: ProvisioningTask, noun: string }) => <span>Creating {noun}: {task.entityId}</span>,
}))
vi.mock("@/pages/runners/components/RunnerRow/RunnerRow.tsx", () => ({
    default: ({runner}: { runner: { name?: string } }) => <span>runner {runner.name}</span>,
}))
vi.mock("@/dialogs/CreateServiceDialog/CreateServiceDialog.tsx", () => ({default: () => null}))

type RunnersQuery = ReturnType<typeof useListRunnersQuery>
type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>

function renderPage(
    data: { runners?: { name: string }[], provisioning?: ProvisioningTask[] },
    isLoading = false,
) {
    vi.mocked(useListRunnersQuery).mockReturnValue({data, isLoading} as Partial<RunnersQuery> as RunnersQuery)
    vi.mocked(useDialog).mockReturnValue({OpenDialog: vi.fn()} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({catchGrpc: vi.fn()} as Partial<Toaster> as Toaster)

    render(<RunnersPage/>)
}

const RUNNING_TASK: ProvisioningTask = {taskId: "1", entityId: "build", status: ProvisioningTaskStatus.RUNNING}

afterEach(() => {
    vi.clearAllMocks()
})

describe("RunnersPage", () => {
    it("shows a provisioning row and no empty state when only a create task exists", () => {
        renderPage({runners: [], provisioning: [RUNNING_TASK]})

        expect(screen.getByText("Creating runner: build")).toBeInTheDocument()
        expect(screen.queryByText("No runners on this node.")).not.toBeInTheDocument()
    })

    it("shows the count skeleton instead of a zero count while loading", () => {
        renderPage({}, true)

        expect(screen.queryByText("0 runners")).not.toBeInTheDocument()
        expect(screen.getByText("runners")).toHaveAttribute("aria-busy", "true")
    })

    it("shows the runner count once loaded", () => {
        renderPage({runners: [{name: "a"}, {name: "b"}], provisioning: []})

        expect(screen.getByText("2 runners")).toBeInTheDocument()
    })

    it("shows the empty state when there are no runners and no tasks", () => {
        renderPage({runners: [], provisioning: []})

        expect(screen.getByText("No runners on this node.")).toBeInTheDocument()
    })

    it("hides the provider-prefixed row of a runner that is still being created", () => {
        renderPage({
            runners: [{name: "github_runner_build"}, {name: "gitlab_runner_build"}, {name: "github_runner_deploy"}],
            provisioning: [RUNNING_TASK],
        })

        expect(screen.queryByText("runner github_runner_build")).not.toBeInTheDocument()
        expect(screen.queryByText("runner gitlab_runner_build")).not.toBeInTheDocument()
        expect(screen.getByText("runner github_runner_deploy")).toBeInTheDocument()
        expect(screen.getByText("3 runners")).toBeInTheDocument()
    })
})
