import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"

import {queryClient} from "@/app/queryClient.ts"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {ListEnvironmentsQuery} from "@/processes/queries/control_plane.ts"
import {useListServicesQuery} from "@/processes/queries/services.ts"
import {CreatePgInstanceMutation} from "@/processes/queries/pg_instances.ts"
import Button from "@/components/base/Button.tsx"
import PostgresScreen from "@/dialogs/CreateServiceDialog/screens/PostgresScreen/PostgresScreen.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/control_plane.ts", () => ({ListEnvironmentsQuery: vi.fn()}))
vi.mock("@/processes/queries/services.ts", () => ({useListServicesQuery: vi.fn()}))
vi.mock("@/processes/queries/pg_instances.ts", () => ({
    CreatePgInstanceMutation: vi.fn(),
    PG_INSTANCES_QUERY_KEY: ["pg-instances"],
}))
vi.mock("@/widgets/TaskProgressScreen/TaskProgressScreen.tsx", () => ({
    default: ({start, onSuccess}: { start(): Promise<unknown>, onSuccess?(): void }) => (
        <>
            <Button variant="primary" onClick={() => start()}>start task</Button>
            <Button variant="primary" onClick={() => onSuccess?.()}>finish task</Button>
        </>
    ),
}))

type PgMutation = ReturnType<typeof CreatePgInstanceMutation>
type EnvironmentsQuery = ReturnType<typeof ListEnvironmentsQuery>
type ServicesQuery = ReturnType<typeof useListServicesQuery>
type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>

function renderScreen() {
    const mutateAsync = vi.fn().mockResolvedValue({entityId: "orders", action: "create_pg_instance"})
    const bake = vi.fn()
    vi.mocked(CreatePgInstanceMutation).mockReturnValue({mutateAsync} as Partial<PgMutation> as PgMutation)
    vi.mocked(ListEnvironmentsQuery).mockReturnValue(
        {data: {environments: []}} as Partial<EnvironmentsQuery> as EnvironmentsQuery,
    )
    vi.mocked(useListServicesQuery).mockReturnValue({data: {services: []}} as Partial<ServicesQuery> as ServicesQuery)
    vi.mocked(useDialog).mockReturnValue({CloseDialog: vi.fn()} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({bake, catchGrpc: vi.fn()} as Partial<Toaster> as Toaster)

    render(<PostgresScreen onBusyChange={vi.fn()}/>)
    return {mutateAsync, bake}
}

function typeName(value: string) {
    fireEvent.change(screen.getByText("Name").previousElementSibling as HTMLInputElement, {target: {value}})
}

afterEach(() => {
    vi.restoreAllMocks()
    vi.clearAllMocks()
})

describe("PostgresScreen", () => {
    it("does not call the create RPC before the form is submitted", () => {
        const {mutateAsync} = renderScreen()
        typeName("orders")

        expect(mutateAsync).not.toHaveBeenCalled()
    })

    it("shows the task progress screen and enqueues the request when Create is clicked", async () => {
        const {mutateAsync} = renderScreen()
        typeName(" orders ")

        fireEvent.click(screen.getByRole("button", {name: "Create"}))
        fireEvent.click(screen.getByText("start task"))

        await waitFor(() => expect(mutateAsync).toHaveBeenCalledTimes(1))
        expect(mutateAsync.mock.calls[0][0]).toMatchObject({name: "orders", box: "small"})
    })

    it("refreshes the lists and bakes a toast when the task succeeds", () => {
        const invalidate = vi.spyOn(queryClient, "invalidateQueries").mockResolvedValue()
        const {bake} = renderScreen()
        typeName("orders")

        fireEvent.click(screen.getByRole("button", {name: "Create"}))
        fireEvent.click(screen.getByText("finish task"))

        expect(invalidate).toHaveBeenCalledWith({queryKey: ["pg-instances"]})
        expect(invalidate).toHaveBeenCalledWith({queryKey: ["services"]})
        expect(bake).toHaveBeenCalledWith({title: "Database created", description: "orders", level: "Info"})
    })
})
