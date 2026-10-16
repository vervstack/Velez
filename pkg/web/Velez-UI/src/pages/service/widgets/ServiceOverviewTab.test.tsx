import {afterEach, describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"

import ServiceOverviewTab from "@/pages/service/widgets/ServiceOverviewTab.tsx"
import {
    GetServiceByNameQuery,
    ListDeploymentsByServiceNameQuery,
    useGetServiceAboutQuery,
} from "@/processes/queries/services.ts"

vi.mock("@/processes/queries/services.ts", () => ({
    GetServiceByNameQuery: vi.fn(),
    ListDeploymentsByServiceNameQuery: vi.fn(),
    useGetServiceAboutQuery: vi.fn(),
}))
vi.mock("@/widgets/service/PgConnectionSection/PgConnectionSection.tsx", () => ({default: () => null}))
vi.mock("@/widgets/service/ServiceHero/ServiceHero.tsx", () => ({default: () => null}))
vi.mock("@/widgets/service/ObservabilityTools/ObservabilityTools.tsx", () => ({default: () => null}))
vi.mock("@/widgets/service/ResourcesSection/ResourcesSection.tsx", () => ({default: () => null}))
vi.mock("@/widgets/service/ServiceGraph/ServiceGraph.tsx", () => ({default: () => null}))
vi.mock("@/widgets/service/DeploymentHistory/DeploymentHistory.tsx", () => ({default: () => null}))
vi.mock("@/widgets/service/Vervonomicon/Vervonomicon.tsx", () => ({default: () => null}))
vi.mock("@/widgets/S3Credentials/S3Credentials.tsx", () => ({
    default: ({instanceName}: {instanceName: string}) => <div>S3 credentials for {instanceName}</div>,
}))

type ServiceQuery = ReturnType<typeof GetServiceByNameQuery>
type DeploymentsQuery = ReturnType<typeof ListDeploymentsByServiceNameQuery>
type AboutQuery = ReturnType<typeof useGetServiceAboutQuery>

function renderTab(serviceName: string) {
    vi.mocked(GetServiceByNameQuery).mockReturnValue({data: {name: serviceName}} as Partial<ServiceQuery> as ServiceQuery)
    vi.mocked(ListDeploymentsByServiceNameQuery).mockReturnValue(
        {data: {deployments: []}} as Partial<DeploymentsQuery> as DeploymentsQuery
    )
    vi.mocked(useGetServiceAboutQuery).mockReturnValue({data: undefined} as Partial<AboutQuery> as AboutQuery)

    render(<ServiceOverviewTab serviceName={serviceName}/>)
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("ServiceOverviewTab", () => {
    it("shows the S3 credentials block for an s3_ service", () => {
        renderTab("s3_tstd")

        expect(screen.getByText("S3 credentials for tstd")).toBeTruthy()
    })

    it("hides the S3 credentials block for a non-S3 service", () => {
        renderTab("postgres_main")

        expect(screen.queryByText(/S3 credentials for/)).toBeNull()
    })
})
