import {afterEach, describe, expect, it, vi, type Mock} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import type {DockerContainer} from "@/app/api/velez"
import {IsStatefullModeEnabled} from "@/processes/queries/control_plane.ts"
import {useImageVersionsQuery} from "@/processes/queries/containers.ts"
import AdoptForm from "@/dialogs/CreateServiceDialog/components/AdoptForm/AdoptForm.tsx"
import {AdoptGroup, AdoptGroupContext} from "@/dialogs/CreateServiceDialog/hooks/AdoptGroup.ts"
import type {OnboardFormState} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"
import {initialRoles, pickService} from "@/dialogs/CreateServiceDialog/processes/registerRoles.ts"

vi.mock("@/processes/queries/control_plane.ts", () => ({IsStatefullModeEnabled: vi.fn()}))
vi.mock("@/processes/queries/containers.ts", () => ({useImageVersionsQuery: vi.fn()}))

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
    group?: AdoptGroup
}

function renderForm({isClusterMode, container = CONTAINER, group}: RenderOptions) {
    vi.mocked(IsStatefullModeEnabled).mockReturnValue(isClusterMode)
    vi.mocked(useImageVersionsQuery).mockReturnValue(
        {isLoading: false, isError: false, data: {tags: []}} as Partial<ReturnType<typeof useImageVersionsQuery>> as
            ReturnType<typeof useImageVersionsQuery>
    )

    const onFormChange = vi.fn<(form: OnboardFormState) => void>()
    render(
        <AdoptGroupContext.Provider value={group ?? null}>
            <AdoptForm container={container} pattern="generic" onFormChange={onFormChange}/>
        </AdoptGroupContext.Provider>
    )
    return {onFormChange}
}

function lastForm(onFormChange: Mock<(form: OnboardFormState) => void>): OnboardFormState {
    return onFormChange.mock.calls[onFormChange.mock.calls.length - 1][0]
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

    it("reports no request when the new host port is invalid", () => {
        const {onFormChange} = renderForm({isClusterMode: false})

        fireEvent.change(screen.getByDisplayValue("15432"), {target: {value: "70000"}})

        expect(lastForm(onFormChange).request).toBeNull()
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

    it("reports a recreation with published ports on a single-node setup", () => {
        const {onFormChange} = renderForm({isClusterMode: false})

        expect(lastForm(onFormChange)).toMatchObject({isRecreating: true, hasPublishedPorts: true})
        expect(lastForm(onFormChange).request).toMatchObject({containerId: "c1", serviceName: "db"})
    })

    it("reports no recreation in cluster mode", () => {
        const {onFormChange} = renderForm({isClusterMode: true})

        expect(lastForm(onFormChange)).toMatchObject({isRecreating: false, hasPublishedPorts: true})
    })

    it("reports no request while the service name is empty", () => {
        const {onFormChange} = renderForm({isClusterMode: true})

        fireEvent.change(screen.getByDisplayValue("db"), {target: {value: ""}})

        expect(lastForm(onFormChange).request).toBeNull()
    })

    it("hides the image tag, links and port options once another container becomes the service", () => {
        const group: AdoptGroup = {
            containers: [CONTAINER, {id: "c2", name: "app", imageName: "nginx:1"}],
            originalRootId: "c1",
            roles: pickService(initialRoles("c1"), "c2"),
            serviceName: "app",
            onServiceNameChange: vi.fn(),
        }

        const {onFormChange} = renderForm({isClusterMode: false, group})

        expect(screen.queryByText("Host new")).not.toBeInTheDocument()
        expect(screen.queryByText("/srv/pg → /var/lib/postgresql/data")).not.toBeInTheDocument()
        expect(screen.queryByText(/publishes no ports/)).not.toBeInTheDocument()
        expect(lastForm(onFormChange).request).not.toBeNull()
    })

    it("lists the links of the service container and sends them in cluster mode", () => {
        const group: AdoptGroup = {
            containers: [
                CONTAINER,
                {
                    id: "c2",
                    name: "app",
                    imageName: "nginx:1",
                    mounts: [{type: "bind", source: "/srv/web", destination: "/etc/nginx"}],
                },
            ],
            originalRootId: "c1",
            roles: pickService(initialRoles("c1"), "c2"),
            serviceName: "app",
            onServiceNameChange: vi.fn(),
        }
        const {onFormChange} = renderForm({isClusterMode: true, group})

        expect(screen.getByText("/srv/web → /etc/nginx")).toBeInTheDocument()
        expect(screen.queryByText("/srv/pg → /var/lib/postgresql/data")).not.toBeInTheDocument()
        expect(lastForm(onFormChange).request).toHaveProperty(
            "bindMountLinks",
            [{source: "/srv/web", volumeName: "app_etc_nginx"}],
        )
    })

    it("reads and edits the service name through the group", () => {
        const onServiceNameChange = vi.fn()
        const group: AdoptGroup = {
            containers: [CONTAINER],
            originalRootId: "c1",
            roles: initialRoles("c1"),
            serviceName: "from-group",
            onServiceNameChange,
        }
        renderForm({isClusterMode: true, group})

        fireEvent.change(screen.getByDisplayValue("from-group"), {target: {value: "renamed"}})

        expect(onServiceNameChange).toHaveBeenCalledWith("renamed")
    })

    it("still reports a recreation in cluster mode when the network root is switched", () => {
        const group: AdoptGroup = {
            containers: [CONTAINER, {id: "c2", name: "app", imageName: "nginx:1"}],
            originalRootId: "c1",
            roles: {rootId: "c2"},
            serviceName: "app",
            onServiceNameChange: vi.fn(),
        }
        const {onFormChange} = renderForm({isClusterMode: true, group})

        expect(lastForm(onFormChange).isRecreating).toBe(true)
    })
})
