import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"
import {useNavigate} from "react-router-dom"

import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {CreateNetworkMutation, useNetworkStatusQuery} from "@/processes/queries/networks.ts"
import CreateNetworkDialog from "@/dialogs/CreateNetworkDialog/CreateNetworkDialog.tsx"

vi.mock("react-router-dom", () => ({useNavigate: vi.fn()}))
vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/networks.ts", () => ({
    CreateNetworkMutation: vi.fn(),
    useNetworkStatusQuery: vi.fn(),
}))

type CreateMutation = ReturnType<typeof CreateNetworkMutation>
type StatusQuery = ReturnType<typeof useNetworkStatusQuery>
type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>

function inputFor(label: string) {
    return screen.getByText(label).previousElementSibling as HTMLInputElement
}

function renderDialog(status: StatusQuery["data"] = {}) {
    const mutateAsync = vi.fn().mockResolvedValue({})
    const CloseDialog = vi.fn()
    const catchGrpc = vi.fn()
    const navigate = vi.fn()
    useEnvironmentStore.setState({selectedEnvironment: "PROD"})
    vi.mocked(CreateNetworkMutation).mockReturnValue({mutateAsync} as Partial<CreateMutation> as CreateMutation)
    vi.mocked(useNetworkStatusQuery).mockReturnValue({data: status} as Partial<StatusQuery> as StatusQuery)
    vi.mocked(useDialog).mockReturnValue({CloseDialog} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({bake: vi.fn(), catchGrpc} as Partial<Toaster> as Toaster)
    vi.mocked(useNavigate).mockReturnValue(navigate)

    render(<CreateNetworkDialog/>)
    return {mutateAsync, CloseDialog, catchGrpc, navigate}
}

function typeName(name: string) {
    fireEvent.change(inputFor("Name"), {target: {value: name}})
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("CreateNetworkDialog", () => {
    it("disables Create while the name is empty or blank", () => {
        renderDialog()

        expect(screen.getByRole("button", {name: "Create"})).toBeDisabled()

        typeName("   ")
        expect(screen.getByRole("button", {name: "Create"})).toBeDisabled()

        typeName("backend")
        expect(screen.getByRole("button", {name: "Create"})).toBeEnabled()
    })

    it("creates a network with container traffic enabled by default", async () => {
        const {mutateAsync, CloseDialog} = renderDialog()
        typeName(" backend ")

        fireEvent.click(screen.getByRole("button", {name: "Create"}))

        await waitFor(() => expect(mutateAsync).toHaveBeenCalledWith(
            {name: "backend", environment: "PROD", isInternal: false, isIccEnabled: true},
        ))
        await waitFor(() => expect(CloseDialog).toHaveBeenCalledTimes(1))
    })

    it("sends isIccEnabled false and isInternal true when both options are checked", async () => {
        const {mutateAsync} = renderDialog()
        typeName("backend")
        fireEvent.click(screen.getByRole("checkbox", {name: /Internal/}))
        fireEvent.click(screen.getByRole("checkbox", {name: /Isolate containers/}))

        fireEvent.click(screen.getByRole("button", {name: "Create"}))

        await waitFor(() => expect(mutateAsync).toHaveBeenCalledWith(
            {name: "backend", environment: "PROD", isInternal: true, isIccEnabled: false},
        ))
    })

    it("routes a failed create through the toaster and keeps the dialog open", async () => {
        const failure = new Error("create failed")
        const {mutateAsync, catchGrpc, CloseDialog} = renderDialog()
        mutateAsync.mockRejectedValue(failure)
        typeName("backend")

        fireEvent.click(screen.getByRole("button", {name: "Create"}))

        await waitFor(() => expect(catchGrpc).toHaveBeenCalledWith(failure))
        expect(CloseDialog).not.toHaveBeenCalled()
    })

    it("shows the cluster banner and navigates to VCN from it", () => {
        const {navigate, CloseDialog} = renderDialog({isClusterMode: true, isVcnConnected: false})

        fireEvent.click(screen.getByRole("button", {name: "Set up VCN"}))

        expect(CloseDialog).toHaveBeenCalledTimes(1)
        expect(navigate).toHaveBeenCalledWith("/vcn")
    })

    it("hides the cluster banner when VCN is connected", () => {
        renderDialog({isClusterMode: true, isVcnConnected: true})

        expect(screen.queryByRole("alert")).not.toBeInTheDocument()
    })
})
