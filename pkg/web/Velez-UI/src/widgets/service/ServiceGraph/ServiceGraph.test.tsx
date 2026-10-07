import {afterEach, describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"

import type {ServiceGraphData, ServiceSidecarView} from "@/model/service_page/ServicePageModel"
import {useGetServiceGraphQuery} from "@/processes/queries/services"
import ServiceGraph from "@/widgets/service/ServiceGraph/ServiceGraph.tsx"

vi.mock("@/processes/queries/services", () => ({useGetServiceGraphQuery: vi.fn()}))

type GraphQuery = ReturnType<typeof useGetServiceGraphQuery>

function renderGraph(sidecars?: ServiceSidecarView[], graph?: Partial<ServiceGraphData>) {
    const data: ServiceGraphData = {
        incoming: [{id: "caller-svc", kind: "service", proto: "grpc", rate: "1 rps"}],
        outgoing: [{id: "postgres-db", kind: "resource", proto: "tcp", rate: "2 rps"}],
        ...graph,
    }
    vi.mocked(useGetServiceGraphQuery).mockReturnValue(
        {data, isLoading: false, isError: false} as Partial<GraphQuery> as GraphQuery,
    )

    return render(<ServiceGraph serviceName="s3_artel" sidecars={sidecars}/>)
}

function webUiSidecar(): ServiceSidecarView {
    return {containerId: "c1", name: "s3_artel_web_ui", imageName: "garage-webui"}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("ServiceGraph", () => {
    it("draws the sidecar and the group caption when sidecars are provided", () => {
        renderGraph([webUiSidecar()])

        expect(screen.getByText("s3_artel_web_ui")).toBeTruthy()
        expect(screen.getByText("service + sidecars")).toBeTruthy()
        expect(screen.getByText("sidecar (same service)")).toBeTruthy()
    })

    it("draws no sidecar group when there are no sidecars", () => {
        renderGraph(undefined)

        expect(screen.queryByText("service + sidecars")).toBeNull()
        expect(screen.queryByText("sidecar (same service)")).toBeNull()
    })

    it("keeps drawing incoming and outgoing nodes next to the sidecar group", () => {
        renderGraph([webUiSidecar()])

        expect(screen.getByText("caller-svc")).toBeTruthy()
        expect(screen.getByText("postgres-db")).toBeTruthy()
    })

    it("shows the empty state when there are no edges and no sidecars", () => {
        renderGraph([], {incoming: [], outgoing: []})

        expect(screen.getByText("No graph data available")).toBeTruthy()
    })

    it("still draws the graph when the service has only sidecars", () => {
        renderGraph([webUiSidecar()], {incoming: [], outgoing: []})

        expect(screen.getByText("service + sidecars")).toBeTruthy()
    })
})
