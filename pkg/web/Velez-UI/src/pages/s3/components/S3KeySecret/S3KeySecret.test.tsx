import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"
import {QueryClient, QueryClientProvider} from "@tanstack/react-query"

import S3KeySecret from "@/pages/s3/components/S3KeySecret/S3KeySecret.tsx"

const getS3KeyCredentials = vi.fn((_instanceName: string, _accessKeyId: string) =>
    Promise.resolve({accessKeyId: "GK1", secretAccessKey: "s3cr3t"}))

vi.mock("@/processes/api/s3aas", () => ({
    s3Service: {
        getS3KeyCredentials: (instanceName: string, accessKeyId: string) =>
            getS3KeyCredentials(instanceName, accessKeyId),
    },
}))

afterEach(() => {
    vi.clearAllMocks()
})

function renderSecret() {
    const queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}})

    return render(
        <QueryClientProvider client={queryClient}>
            <S3KeySecret instanceName="s3-main" accessKeyId="GK1"/>
        </QueryClientProvider>
    )
}

describe("S3KeySecret", () => {
    it("shows the masked secret and does not fetch it before reveal is clicked", () => {
        renderSecret()

        expect(screen.getByText("••••••••••••")).toBeInTheDocument()
        expect(getS3KeyCredentials).not.toHaveBeenCalled()
    })

    it("fetches and reveals the real secret when Reveal is clicked", async () => {
        renderSecret()

        fireEvent.click(screen.getByText("Reveal"))

        await waitFor(() => expect(getS3KeyCredentials).toHaveBeenCalledWith("s3-main", "GK1"))
        await waitFor(() => expect(screen.getByText("s3cr3t")).toBeInTheDocument())
        expect(screen.getByText("Hide")).toBeInTheDocument()
    })

    it("masks the secret again when Hide is clicked", async () => {
        renderSecret()

        fireEvent.click(screen.getByText("Reveal"))
        await waitFor(() => expect(screen.getByText("s3cr3t")).toBeInTheDocument())
        fireEvent.click(screen.getByText("Hide"))

        expect(screen.queryByText("s3cr3t")).not.toBeInTheDocument()
        expect(screen.getByText("••••••••••••")).toBeInTheDocument()
    })

    it("copies the secret, fetching it first if not already loaded", async () => {
        const writeText = vi.fn().mockResolvedValue(undefined)
        Object.assign(navigator, {clipboard: {writeText}})

        renderSecret()

        fireEvent.click(screen.getByText("Copy"))

        await waitFor(() => expect(getS3KeyCredentials).toHaveBeenCalledWith("s3-main", "GK1"))
        await waitFor(() => expect(writeText).toHaveBeenCalledWith("s3cr3t"))
    })
})
