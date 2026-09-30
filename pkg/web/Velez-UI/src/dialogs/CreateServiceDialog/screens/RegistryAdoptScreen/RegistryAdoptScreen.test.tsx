import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import type {DockerContainer} from "@/app/api/velez"
import {useGetContainerQuery} from "@/processes/queries/containers.ts"
import RegistryAdoptScreen from "@/dialogs/CreateServiceDialog/screens/RegistryAdoptScreen/RegistryAdoptScreen.tsx"

vi.mock("@/processes/queries/containers.ts", () => ({useGetContainerQuery: vi.fn()}))
vi.mock("@/dialogs/CreateServiceDialog/components/AdoptForm/AdoptForm.tsx", () => ({
    default: ({isRegistryLoginRequired}: {isRegistryLoginRequired: boolean}) => (
        <span>{isRegistryLoginRequired ? "login required" : "login not required"}</span>
    ),
}))

type QueryResult = ReturnType<typeof useGetContainerQuery>

function renderScreen(query: Partial<QueryResult>) {
    vi.mocked(useGetContainerQuery).mockReturnValue(query as Partial<QueryResult> as QueryResult)

    const container: DockerContainer = {id: "c1", name: "registry"}
    render(<RegistryAdoptScreen container={container} onBusyChange={vi.fn()}/>)
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("RegistryAdoptScreen", () => {
    it("offers a retry when the container fails to load", () => {
        const refetch = vi.fn()
        renderScreen({isError: true, refetch} as Partial<QueryResult>)

        fireEvent.click(screen.getByText("Retry"))

        expect(screen.getByText("Failed to load the container.")).toBeInTheDocument()
        expect(refetch).toHaveBeenCalledTimes(1)
    })

    it("asks for a login when the container env configures auth", () => {
        renderScreen({data: {env: {REGISTRY_AUTH: "htpasswd"}}})

        expect(screen.getByText("login required")).toBeInTheDocument()
    })

    it("asks for nothing when the container env has no auth", () => {
        renderScreen({data: {env: {}}})

        expect(screen.getByText("login not required")).toBeInTheDocument()
    })
})
