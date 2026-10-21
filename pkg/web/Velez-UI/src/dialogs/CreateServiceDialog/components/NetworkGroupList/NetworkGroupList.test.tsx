import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {ValueOrigin, type DockerContainer} from "@/app/api/velez"
import {useRegisterPreviewQuery} from "@/processes/queries/containers.ts"
import NetworkGroupList from "@/dialogs/CreateServiceDialog/components/NetworkGroupList/NetworkGroupList.tsx"
import type {AdoptGroup} from "@/dialogs/CreateServiceDialog/hooks/AdoptGroup.ts"
import {initialRoles, pickService} from "@/dialogs/CreateServiceDialog/processes/registerRoles.ts"

vi.mock("@/processes/queries/containers.ts", () => ({useRegisterPreviewQuery: vi.fn()}))

const ROOT: DockerContainer = {id: "root", name: "vpn", imageName: "tailscale:1"}
const MEMBER: DockerContainer = {id: "app", name: "app", imageName: "nginx:1"}

const RESPONSE = {
    containers: [
        {
            containerId: "root",
            env: [{key: "VERV_SERVICE", value: "web", origin: ValueOrigin.ADDED}],
        },
        {
            containerId: "app",
            env: [{key: "KEPT", value: "yes", origin: ValueOrigin.TRANSFERRED}],
        },
    ],
}

function mockQuery(overrides: Partial<ReturnType<typeof useRegisterPreviewQuery>> = {}) {
    vi.mocked(useRegisterPreviewQuery).mockReturnValue({
        data: RESPONSE,
        isLoading: false,
        isError: false,
        refetch: vi.fn(),
        ...overrides,
    } as Partial<ReturnType<typeof useRegisterPreviewQuery>> as ReturnType<typeof useRegisterPreviewQuery>)
}

function renderList(group: Partial<AdoptGroup> = {}) {
    const onPickService = vi.fn()
    const onRequestRoot = vi.fn()
    const fullGroup: AdoptGroup = {
        containers: [ROOT, MEMBER],
        originalRootId: "root",
        roles: initialRoles("root"),
        serviceName: "web",
        onServiceNameChange: vi.fn(),
        ...group,
    }

    render(
        <NetworkGroupList
            group={fullGroup}
            pattern="generic"
            onPickService={onPickService}
            onRequestRoot={onRequestRoot}
        />
    )
    return {onPickService, onRequestRoot}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("NetworkGroupList", () => {
    it("explains the colours in a legend above the items", () => {
        mockQuery()
        renderList()

        expect(screen.getByText("Transferred from the container")).toBeInTheDocument()
        expect(screen.getByText("Added by onboarding")).toBeInTheDocument()
    })

    it("builds the preview request from the group and the chosen roles", () => {
        mockQuery()
        renderList({roles: pickService(initialRoles("root"), "app")})

        expect(useRegisterPreviewQuery).toHaveBeenCalledWith(expect.objectContaining({
            containerId: "root",
            serviceName: "web",
            serviceContainerId: "app",
            generic: {},
        }))
    })

    it("shows each item its own preview entry", () => {
        mockQuery()
        renderList()

        fireEvent.click(screen.getByRole("button", {name: /^app/}))

        expect(screen.getByText("KEPT")).toBeInTheDocument()
        expect(screen.getAllByText("added")).toHaveLength(1)
    })

    it("asks for a service name instead of a preview while the name is empty", () => {
        mockQuery()
        renderList({serviceName: " "})

        expect(useRegisterPreviewQuery).toHaveBeenCalledWith(null)
        expect(screen.getByText("Enter a service name to preview the result.")).toBeInTheDocument()
    })

    it("notes that only the generic pattern is available once the service moves", () => {
        mockQuery()
        renderList({roles: pickService(initialRoles("root"), "app")})

        expect(screen.getByText(/only the Generic pattern is available/)).toBeInTheDocument()
    })

    it("describes a non-service container as a resource of the named service", () => {
        mockQuery()
        renderList()
        fireEvent.click(screen.getByRole("button", {name: /^app/}))

        expect(screen.getByText("Resource of web")).toBeInTheDocument()
        expect(screen.getByText("Generic container")).toBeInTheDocument()
    })

    it("hands the picked container to the callbacks", () => {
        mockQuery()
        const {onPickService, onRequestRoot} = renderList()
        fireEvent.click(screen.getByRole("button", {name: /^app/}))

        fireEvent.click(screen.getByRole("button", {name: "Make service"}))
        fireEvent.click(screen.getByRole("button", {name: "Make network root"}))

        expect(onPickService).toHaveBeenCalledWith("app")
        expect(onRequestRoot).toHaveBeenCalledWith("app")
    })
})
