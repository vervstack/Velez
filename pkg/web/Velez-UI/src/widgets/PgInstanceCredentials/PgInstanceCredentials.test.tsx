import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"
import {QueryClient, QueryClientProvider} from "@tanstack/react-query"

import type {GetPgInstanceCredentialsResponse} from "@/app/api/velez/pgaas_api.pb.ts"
import PgInstanceCredentials from "@/widgets/PgInstanceCredentials/PgInstanceCredentials.tsx"

const baseCredentials = {
    dbName: "app",
    username: "app_user",
    password: "s3cr3t",
    dsn: "postgres://dsn",
    inNetwork: {dsn: "postgres://app:s3cr3t@pg:5432/app", maskedDsn: "postgres://app:***@pg:5432/app"},
    external: {dsn: "postgres://app:s3cr3t@host:18714/app", maskedDsn: "postgres://app:***@host:18714/app"},
}

const getPgInstanceCredentials = vi.fn((_name: string): Promise<GetPgInstanceCredentialsResponse> =>
    Promise.resolve(baseCredentials))

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
    it("keeps the password masked until Reveal is clicked", async () => {
        renderCredentials()

        await waitFor(() => expect(screen.getByText("postgres://app:***@pg:5432/app")).toBeInTheDocument())
        expect(screen.getByText("••••••••••••")).toBeInTheDocument()
        expect(screen.queryByText("s3cr3t")).not.toBeInTheDocument()

        fireEvent.click(screen.getByText("Reveal"))

        expect(screen.getByText("s3cr3t")).toBeInTheDocument()
        expect(getPgInstanceCredentials).toHaveBeenCalledTimes(1)
    })

    it("shows the masked in-network and external connection strings without revealing", async () => {
        renderCredentials()

        await waitFor(() => expect(screen.getByText("postgres://app:***@pg:5432/app")).toBeInTheDocument())
        expect(screen.getByText("In-network")).toBeInTheDocument()
        expect(screen.getByText("External")).toBeInTheDocument()
        expect(screen.getByText("postgres://app:***@host:18714/app")).toBeInTheDocument()
    })

    it("hides the external row when the port is not published", async () => {
        getPgInstanceCredentials.mockResolvedValueOnce({...baseCredentials, external: undefined})

        renderCredentials()

        await waitFor(() => expect(screen.getByText("postgres://app:***@pg:5432/app")).toBeInTheDocument())
        expect(screen.queryByText("External")).not.toBeInTheDocument()
    })

    it("copies the real DSN of the clicked connection", async () => {
        const writeText = vi.fn().mockResolvedValue(undefined)
        Object.assign(navigator, {clipboard: {writeText}})

        renderCredentials()
        await waitFor(() => expect(screen.getByText("postgres://app:***@pg:5432/app")).toBeInTheDocument())

        const copyButtons = screen.getAllByRole("button", {name: "Copy"})
        fireEvent.click(copyButtons[1])

        await waitFor(() => expect(writeText).toHaveBeenCalledWith("postgres://app:s3cr3t@host:18714/app"))
    })
})
