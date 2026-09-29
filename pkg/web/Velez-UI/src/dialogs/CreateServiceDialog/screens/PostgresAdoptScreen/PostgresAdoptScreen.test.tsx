import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import type {DockerContainer} from "@/app/api/velez"
import {useGetContainerQuery} from "@/processes/queries/containers.ts"
import PostgresAdoptScreen from "@/dialogs/CreateServiceDialog/screens/PostgresAdoptScreen/PostgresAdoptScreen.tsx"

vi.mock("@/processes/queries/containers.ts", () => ({useGetContainerQuery: vi.fn()}))
vi.mock("@/dialogs/CreateServiceDialog/components/AdoptForm/AdoptForm.tsx", () => ({
    default: ({isPgLoginRequired}: {isPgLoginRequired: boolean}) => (
        <span>{isPgLoginRequired ? "login required" : "login not required"}</span>
    ),
}))

type QueryResult = ReturnType<typeof useGetContainerQuery>

function renderScreen(query: Partial<QueryResult>) {
    vi.mocked(useGetContainerQuery).mockReturnValue(query as Partial<QueryResult> as QueryResult)

    const container: DockerContainer = {id: "c1", name: "db"}
    render(<PostgresAdoptScreen container={container} onBusyChange={vi.fn()}/>)
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("PostgresAdoptScreen", () => {
    it("shows no form and no error while the container loads", () => {
        renderScreen({isLoading: true})

        expect(screen.queryByText("login required")).not.toBeInTheDocument()
        expect(screen.queryByText("Failed to load the container.")).not.toBeInTheDocument()
    })

    it("offers a retry when the container fails to load", () => {
        const refetch = vi.fn()
        renderScreen({isError: true, refetch} as Partial<QueryResult>)

        fireEvent.click(screen.getByText("Retry"))

        expect(screen.getByText("Failed to load the container.")).toBeInTheDocument()
        expect(refetch).toHaveBeenCalledTimes(1)
    })

    it("does not ask for a login when the container env carries it", () => {
        renderScreen({data: {env: {POSTGRES_USER: "admin", POSTGRES_PASSWORD: "secret"}}})

        expect(screen.getByText("login not required")).toBeInTheDocument()
    })

    it("asks for a login when the container env lacks it", () => {
        renderScreen({data: {env: {}}})

        expect(screen.getByText("login required")).toBeInTheDocument()
    })
})
