import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import type {DockerContainer} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {IsStatefullModeEnabled} from "@/processes/queries/control_plane.ts"
import {RegisterContainerMutation, useImageVersionsQuery} from "@/processes/queries/containers.ts"
import AdoptForm from "@/dialogs/CreateServiceDialog/components/AdoptForm/AdoptForm.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/processes/queries/control_plane.ts", () => ({IsStatefullModeEnabled: vi.fn()}))
vi.mock("@/processes/queries/containers.ts", () => ({
    CONTAINERS_QUERY_KEY: ["containers"],
    CONTAINER_QUERY_KEY: ["container"],
    RegisterContainerMutation: vi.fn(),
    useImageVersionsQuery: vi.fn(),
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
    vi.mocked(useImageVersionsQuery).mockReturnValue(
        {isLoading: false, isError: false, data: {tags: []}} as Partial<ReturnType<typeof useImageVersionsQuery>> as
            ReturnType<typeof useImageVersionsQuery>
    )
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
    it("warns about a recreation and shows the port mapping table on a single-node setup", () => {
        renderForm({isClusterMode: false})

        expect(screen.getByRole("note")).toHaveTextContent("The container will be recreated")
        expect(screen.getByText("Host new")).toBeInTheDocument()
        expect(screen.getByDisplayValue("15432")).toBeInTheDocument()
        expect(screen.getByText(/Port 15432 is kept: the service is stopped completely/)).toBeInTheDocument()
    })

    it("explains a paused old container when the host port is changed", () => {
        renderForm({isClusterMode: false})

        fireEvent.change(screen.getByDisplayValue("15432"), {target: {value: "15433"}})

        expect(screen.getByText("Port 15432 -> 15433: the old container is paused until you finish onboarding."))
            .toBeInTheDocument()
    })

    it("warns that a container with volumes is stopped before the new one starts", () => {
        renderForm({isClusterMode: false})

        expect(screen.getByText(/This container has volumes: it is stopped \(not paused\)/)).toBeInTheDocument()
    })

    it("disables Register when the new host port is invalid", () => {
        renderForm({isClusterMode: false})

        fireEvent.change(screen.getByDisplayValue("15432"), {target: {value: "70000"}})

        expect(screen.getByRole("button", {name: "Onboard"})).toBeDisabled()
        expect(screen.getByText(/Each host port must be a whole number/)).toBeInTheDocument()
    })

    it("says the container stays running and hides the port table in cluster mode", () => {
        renderForm({isClusterMode: true})

        expect(screen.getByRole("note")).toHaveTextContent("The container stays running")
        expect(screen.queryByText("Host new")).not.toBeInTheDocument()
    })

    it("hides the port options when the container publishes no host ports", () => {
        renderForm({isClusterMode: false, container: {...CONTAINER, ports: []}})

        expect(screen.queryByText("Host new")).not.toBeInTheDocument()
        expect(screen.getByText("This container publishes no ports, so no port mapping is created."))
            .toBeInTheDocument()
    })

    it("does not say the container publishes no ports when it has some", () => {
        renderForm({isClusterMode: false})

        expect(screen.queryByText(/publishes no ports/)).not.toBeInTheDocument()
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

        fireEvent.click(screen.getByRole("button", {name: "Onboard"}))

        expect(screen.getByText("Restart container?")).toBeInTheDocument()
        expect(screen.queryByText("progress screen")).not.toBeInTheDocument()
        expect(onBusyChange).not.toHaveBeenLastCalledWith(true)

        fireEvent.click(screen.getByRole("button", {name: "Onboard and restart"}))

        expect(screen.getByText("progress screen")).toBeInTheDocument()
    })

    it("registers straight away in cluster mode", () => {
        const {onBusyChange} = renderForm({isClusterMode: true})

        fireEvent.click(screen.getByRole("button", {name: "Onboard"}))

        expect(screen.getByText("progress screen")).toBeInTheDocument()
        expect(onBusyChange).toHaveBeenLastCalledWith(true)
    })

    it("disables Register while the service name is empty", () => {
        renderForm({isClusterMode: true})

        fireEvent.change(screen.getByDisplayValue("db"), {target: {value: ""}})

        expect(screen.getByRole("button", {name: "Onboard"})).toBeDisabled()
    })
})
