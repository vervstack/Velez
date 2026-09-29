import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import type {DockerContainer} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {IsStatefullModeEnabled} from "@/processes/queries/control_plane.ts"
import {RegisterContainerMutation} from "@/processes/queries/containers.ts"
import AdoptForm from "@/dialogs/CreateServiceDialog/components/AdoptForm/AdoptForm.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/processes/queries/control_plane.ts", () => ({IsStatefullModeEnabled: vi.fn()}))
vi.mock("@/processes/queries/containers.ts", () => ({
    CONTAINERS_QUERY_KEY: ["containers"],
    CONTAINER_QUERY_KEY: ["container"],
    RegisterContainerMutation: vi.fn(),
}))
vi.mock("@/widgets/TaskProgressScreen/TaskProgressScreen.tsx", () => ({
    default: () => <span>progress screen</span>,
}))

const CONTAINER: DockerContainer = {
    id: "c1",
    name: "db",
    imageName: "postgres:16",
    ports: [{servicePortNumber: 5432, exposedTo: 15432}],
    mounts: [{type: "bind", source: "/srv/pg", destination: "/var/lib/postgresql/data"}],
}

interface RenderOptions {
    isClusterMode: boolean
    container?: DockerContainer
}

function renderForm({isClusterMode, container = CONTAINER}: RenderOptions) {
    vi.mocked(useDialog).mockReturnValue({
        LockClosing: vi.fn(),
        UnlockClosing: vi.fn(),
        CloseDialog: vi.fn(),
    } as Partial<ReturnType<typeof useDialog>> as ReturnType<typeof useDialog>)
    vi.mocked(IsStatefullModeEnabled).mockReturnValue(isClusterMode)
    vi.mocked(RegisterContainerMutation).mockReturnValue(
        {mutateAsync: vi.fn()} as Partial<ReturnType<typeof RegisterContainerMutation>> as
            ReturnType<typeof RegisterContainerMutation>
    )

    const onBusyChange = vi.fn()
    render(<AdoptForm container={container} pattern="generic" onBusyChange={onBusyChange}/>)
    return {onBusyChange}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("AdoptForm", () => {
    it("warns about a restart and offers the port options on a single-node setup", () => {
        renderForm({isClusterMode: false})

        expect(screen.getByRole("note")).toHaveTextContent("The container will be restarted")
        expect(screen.getByText("Keep existing ports (stops the container first → downtime)")).toBeInTheDocument()
        expect(screen.getByText("15432 → 5432/tcp")).toBeInTheDocument()
    })

    it("says the container stays running and hides the port options in cluster mode", () => {
        renderForm({isClusterMode: true})

        expect(screen.getByRole("note")).toHaveTextContent("The container stays running")
        expect(screen.queryByText("Keep existing ports (stops the container first → downtime)"))
            .not.toBeInTheDocument()
    })

    it("hides the port options when the container publishes no host ports", () => {
        renderForm({isClusterMode: false, container: {...CONTAINER, ports: []}})

        expect(screen.queryByText("Ports")).not.toBeInTheDocument()
        expect(screen.getByRole("note")).toBeInTheDocument()
    })

    it("prefills the service name and derives the link volume name from it", () => {
        renderForm({isClusterMode: true})

        expect(screen.getByText("/srv/pg → /var/lib/postgresql/data")).toBeInTheDocument()
        const values = screen.getAllByRole("textbox").map((input) => (input as HTMLInputElement).value)
        expect(values).toContain("db")
        expect(values).toContain("db_var_lib_postgresql_data")
    })

    it("asks for confirmation before registering on a single-node setup", () => {
        const {onBusyChange} = renderForm({isClusterMode: false})

        fireEvent.click(screen.getByRole("button", {name: "Register"}))

        expect(screen.getByText("Restart container?")).toBeInTheDocument()
        expect(screen.queryByText("progress screen")).not.toBeInTheDocument()
        expect(onBusyChange).not.toHaveBeenLastCalledWith(true)

        fireEvent.click(screen.getByRole("button", {name: "Register and restart"}))

        expect(screen.getByText("progress screen")).toBeInTheDocument()
    })

    it("registers straight away in cluster mode", () => {
        const {onBusyChange} = renderForm({isClusterMode: true})

        fireEvent.click(screen.getByRole("button", {name: "Register"}))

        expect(screen.getByText("progress screen")).toBeInTheDocument()
        expect(onBusyChange).toHaveBeenLastCalledWith(true)
    })

    it("disables Register while the service name is empty", () => {
        renderForm({isClusterMode: true})

        fireEvent.change(screen.getByDisplayValue("db"), {target: {value: ""}})

        expect(screen.getByRole("button", {name: "Register"})).toBeDisabled()
    })
})
