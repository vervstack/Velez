import {afterEach, describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"

import {GetS3InstanceCredentialsQuery} from "@/processes/queries/s3.ts"
import S3Credentials from "@/widgets/S3Credentials/S3Credentials.tsx"

vi.mock("@/processes/queries/s3.ts", () => ({
    GetS3InstanceCredentialsQuery: vi.fn(),
}))

type QueryResult = ReturnType<typeof GetS3InstanceCredentialsQuery>
type Credentials = NonNullable<QueryResult["data"]>

const credentials: Partial<Credentials> = {
    adminToken: "admin-token-value",
    s3Endpoint: "https://s3.example.test",
    internalS3Endpoint: "http://s3:3900",
    region: "garage",
    webUiUrl: "https://ui.example.test",
    webUiUsername: "webui-admin",
    webUiPassword: "webui-pass-value",
}

afterEach(() => {
    vi.clearAllMocks()
})

function renderCredentials(data?: Partial<Credentials>) {
    const refetch = vi.fn()
    const query = {
        data: data as Credentials | undefined,
        isFetching: false,
        isError: false,
        refetch,
    } as Partial<QueryResult> as QueryResult

    vi.mocked(GetS3InstanceCredentialsQuery).mockReturnValue(query)

    render(<S3Credentials instanceName="files"/>)
    return {refetch}
}

describe("S3Credentials", () => {
    it("shows Load credentials button before any fetch", () => {
        renderCredentials()

        expect(screen.getByRole("button", {name: "Load credentials"})).toBeInTheDocument()
        expect(screen.queryByText("webui-admin")).not.toBeInTheDocument()
    })

    it("shows credential rows after data arrives, with secrets masked", () => {
        renderCredentials(credentials)

        expect(screen.getByText("Web UI user")).toBeInTheDocument()
        expect(screen.getByText("webui-admin")).toBeInTheDocument()
        expect(screen.getAllByText("••••••••••••")).toHaveLength(2)
        expect(screen.queryByText("webui-pass-value")).not.toBeInTheDocument()
        expect(screen.queryByText("admin-token-value")).not.toBeInTheDocument()
    })
})
