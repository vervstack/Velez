import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"
import {QueryClient, QueryClientProvider} from "@tanstack/react-query"

import PgInstanceCredentials from "@/pages/postgres/components/PgInstanceCredentials/PgInstanceCredentials.tsx"

const getPgInstanceCredentials = vi.fn((_name: string) =>
    Promise.resolve({dbName: "app", username: "app_user", password: "s3cr3t", dsn: "postgres://dsn"}))

vi.mock("@/processes/api/pgaas", () => ({
    pgaasService: {
        getPgInstanceCredentials: (name: string) => getPgInstanceCredentials(name),
    },
}))

afterEach(() => {
    vi.clearAllMocks()
})

function renderCredentials() {
    const queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}})

    return render(
        <QueryClientProvider client={queryClient}>
            <PgInstanceCredentials name="pg-app" dbName="app" username="app_user"/>
        </QueryClientProvider>
    )
}

describe("PgInstanceCredentials", () => {
    it("shows the masked password and does not fetch credentials before reveal is clicked", () => {
        renderCredentials()

        expect(screen.getByText("••••••••••••")).toBeInTheDocument()
        expect(getPgInstanceCredentials).not.toHaveBeenCalled()
    })

    it("fetches and reveals the real password when Reveal is clicked", async () => {
        renderCredentials()

        fireEvent.click(screen.getByText("Reveal"))

        await waitFor(() => expect(getPgInstanceCredentials).toHaveBeenCalledWith("pg-app"))
        await waitFor(() => expect(screen.getByText("s3cr3t")).toBeInTheDocument())
    })

    it("copies the DSN, fetching credentials first if not already loaded", async () => {
        const writeText = vi.fn().mockResolvedValue(undefined)
        Object.assign(navigator, {clipboard: {writeText}})

        renderCredentials()

        fireEvent.click(screen.getByText("Copy DSN"))

        await waitFor(() => expect(getPgInstanceCredentials).toHaveBeenCalledWith("pg-app"))
        await waitFor(() => expect(writeText).toHaveBeenCalledWith("postgres://dsn"))
    })
})
