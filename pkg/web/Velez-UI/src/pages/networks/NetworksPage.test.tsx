import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"
import {useNavigate} from "react-router-dom"

import type {Network, NetworkCapability} from "@/app/api/velez/network_api.pb"
import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {
    DeleteNetworkMutation,
    DisconnectContainerMutation,
    useListNetworksQuery,
    useNetworkStatusQuery,
} from "@/processes/queries/networks.ts"
import NetworksPage from "@/pages/networks/NetworksPage.tsx"
import CreateNetworkDialog from "@/dialogs/CreateNetworkDialog/CreateNetworkDialog.tsx"

vi.mock("react-router-dom", () => ({useNavigate: vi.fn()}))
vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/networks.ts", () => ({
    useListNetworksQuery: vi.fn(),
    useNetworkStatusQuery: vi.fn(),
    DeleteNetworkMutation: vi.fn(),
    DisconnectContainerMutation: vi.fn(),
}))
vi.mock("@/dialogs/CreateNetworkDialog/CreateNetworkDialog.tsx", () => ({default: () => null}))
vi.mock("@/dialogs/AttachContainerDialog/AttachContainerDialog.tsx", () => ({default: () => null}))
vi.mock("@/pages/networks/components/NetworksListSkeleton/NetworksListSkeleton.tsx", () => ({
    default: () => <span>networks skeleton</span>,
}))

type ListQuery = ReturnType<typeof useListNetworksQuery>
type StatusQuery = ReturnType<typeof useNetworkStatusQuery>
const DELETE = "NETWORK_CAPABILITY_DELETE" as NetworkCapability

type DeleteMutation = ReturnType<typeof DeleteNetworkMutation>
type DisconnectMutation = ReturnType<typeof DisconnectContainerMutation>
type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>

function renderPage(list: Partial<ListQuery>, status: StatusQuery["data"] = {}) {
    const refetch = vi.fn()
    const OpenDialog = vi.fn()
    const navigate = vi.fn()
    useEnvironmentStore.setState({selectedEnvironment: "PROD"})
    vi.mocked(useListNetworksQuery).mockReturnValue({refetch, ...list} as Partial<ListQuery> as ListQuery)
    vi.mocked(useNetworkStatusQuery).mockReturnValue({data: status} as Partial<StatusQuery> as StatusQuery)
    vi.mocked(DeleteNetworkMutation).mockReturnValue(
        {mutateAsync: vi.fn()} as Partial<DeleteMutation> as DeleteMutation,
    )
    vi.mocked(DisconnectContainerMutation).mockReturnValue(
        {mutateAsync: vi.fn()} as Partial<DisconnectMutation> as DisconnectMutation,
    )
    vi.mocked(useDialog).mockReturnValue({OpenDialog, CloseDialog: vi.fn()} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({bake: vi.fn(), catchGrpc: vi.fn()} as Partial<Toaster> as Toaster)
    vi.mocked(useNavigate).mockReturnValue(navigate)

    render(<NetworksPage/>)
    return {refetch, OpenDialog, navigate}
}

function loaded(networks: Network[]): Partial<ListQuery> {
    return {data: {networks}}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("NetworksPage", () => {
    it("lists managed networks before foreign ones", () => {
        renderPage(loaded([
            {id: "1", name: "bridge"},
            {id: "2", name: "app-net", isManaged: true, capabilities: [DELETE]},
        ]))

        const names = screen.getAllByText(/^(bridge|app-net)$/).map((el) => el.textContent)
        expect(names).toEqual(["app-net", "bridge"])
    })

    it("shows the skeleton while loading", () => {
        renderPage({isLoading: true})

        expect(screen.getByText("networks skeleton")).toBeInTheDocument()
        expect(screen.queryByText("No networks in this environment.")).not.toBeInTheDocument()
    })

    it("shows the empty state when there are no networks", () => {
        renderPage(loaded([]))

        expect(screen.getByText("No networks in this environment.")).toBeInTheDocument()
    })

    it("shows a retry-capable error state that refetches", () => {
        const {refetch} = renderPage({isError: true})

        expect(screen.getByText("Failed to load networks.")).toBeInTheDocument()
        fireEvent.click(screen.getByText("Retry"))

        expect(refetch).toHaveBeenCalledTimes(1)
    })

    it("shows the cluster banner when cluster mode is on and VCN is not connected", () => {
        const {navigate} = renderPage(loaded([]), {isClusterMode: true, isVcnConnected: false})

        fireEvent.click(screen.getByRole("button", {name: "Set up VCN"}))

        expect(screen.getByRole("alert")).toBeInTheDocument()
        expect(navigate).toHaveBeenCalledWith("/vcn")
    })

    it("hides the cluster banner when VCN is connected", () => {
        renderPage(loaded([]), {isClusterMode: true, isVcnConnected: true})

        expect(screen.queryByRole("alert")).not.toBeInTheDocument()
    })

    it("queries the selected environment without foreign networks by default", () => {
        renderPage(loaded([]))

        expect(useListNetworksQuery).toHaveBeenLastCalledWith("PROD", false)
    })

    it("asks for foreign networks after the toggle is switched on", () => {
        renderPage(loaded([]))

        fireEvent.click(screen.getByRole("switch", {name: "Show all Docker networks"}))

        expect(useListNetworksQuery).toHaveBeenLastCalledWith("PROD", true)
    })

    it("opens CreateNetworkDialog when Create network is clicked", () => {
        const {OpenDialog} = renderPage(loaded([]))

        fireEvent.click(screen.getByRole("button", {name: "Create network"}))

        expect(OpenDialog).toHaveBeenCalledTimes(1)
        expect(OpenDialog.mock.calls[0][0].type).toBe(CreateNetworkDialog)
    })
})
