import {afterEach, describe, expect, it, vi} from "vitest"
import {render} from "@testing-library/react"

import ServiceDetailLayout from "@/pages/service/widgets/ServiceDetailLayout.tsx"
import {useBreadcrumbs} from "@/app/hooks/breadcrumbs/Breadcrumbs.ts"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {GetServiceByNameQuery} from "@/processes/queries/services.ts"

vi.mock("react-router-dom", () => ({useNavigate: () => vi.fn()}))
vi.mock("@/app/hooks/breadcrumbs/Breadcrumbs.ts", () => ({useBreadcrumbs: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/services.ts", () => ({
    GetServiceByNameQuery: vi.fn(),
}))
vi.mock("@/pages/service/widgets/TabStrip.tsx", () => ({default: () => null}))
vi.mock("@/pages/service/widgets/ServiceTagsStrip.tsx", () => ({default: () => null}))
vi.mock("@/pages/service/widgets/ServiceOverviewTab.tsx", () => ({default: () => null}))
vi.mock("@/pages/service/widgets/ServicePageSkeleton.tsx", () => ({default: () => null}))

type Toaster = ReturnType<typeof useToaster>
type ServiceQuery = ReturnType<typeof GetServiceByNameQuery>

function renderLayout(serviceName: string, displayName?: string) {
    const setCrumbs = vi.fn()
    vi.mocked(useBreadcrumbs).mockImplementation(
        ((selector: (s: {setCrumbs: typeof setCrumbs}) => unknown) => selector({setCrumbs})) as typeof useBreadcrumbs
    )
    vi.mocked(useToaster).mockReturnValue({catchGrpc: vi.fn()} as Partial<Toaster> as Toaster)
    vi.mocked(GetServiceByNameQuery).mockReturnValue(
        {
            data: {name: serviceName, displayName},
            isLoading: false,
            isError: false,
        } as Partial<ServiceQuery> as ServiceQuery
    )

    render(<ServiceDetailLayout serviceName={serviceName}/>)
    return {setCrumbs}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("ServiceDetailLayout", () => {
    it("publishes the service display name as the breadcrumb when the service has one", () => {
        const {setCrumbs} = renderLayout("pgaas_main", "main")

        expect(setCrumbs).toHaveBeenCalledWith([
            {label: "services", onClick: expect.any(Function)},
            {label: "main"},
        ])
    })

    it("falls back to the derived runner name when the service has no display name", () => {
        const {setCrumbs} = renderLayout("gitlab_runner_ci")

        expect(setCrumbs).toHaveBeenCalledWith([
            {label: "services", onClick: expect.any(Function)},
            {label: "ci"},
        ])
    })
})
