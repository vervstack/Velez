import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"
import {QueryClient, QueryClientProvider} from "@tanstack/react-query"

import type {PlanTransferResponse, StartTransferRequest} from "@/app/api/velez/transfer_api.pb"
import MoveContainerForm from "@/pages/transfer/components/MoveContainerForm/MoveContainerForm.tsx"

const planTransfer = vi.fn((_containerName: string) => Promise.resolve({} as PlanTransferResponse))
const startTransfer = vi.fn((_req: StartTransferRequest) => Promise.resolve({transferId: "t-1"}))

vi.mock("@/processes/api/transfer", () => ({
    transferService: {
        planTransfer: (containerName: string) => planTransfer(containerName),
        startTransfer: (req: StartTransferRequest) => startTransfer(req),
    },
}))

vi.mock("@/pages/transfer/components/ContainerPicker/ContainerPicker.tsx", () => ({
    default: ({onChange}: {onChange(name: string): void}) => (
        <span role="option" aria-selected="false" onClick={() => onChange("app")}>Pick app</span>
    ),
}))

afterEach(() => {
    vi.clearAllMocks()
})

function renderForm() {
    const queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}})

    return render(
        <QueryClientProvider client={queryClient}>
            <MoveContainerForm/>
        </QueryClientProvider>
    )
}

function fillAndCheck() {
    const [address, token] = screen.getAllByRole("textbox")
    fireEvent.click(screen.getByText("Pick app"))
    fireEvent.change(address, {target: {value: "10.0.0.2:53890"}})
    fireEvent.change(token, {target: {value: "tok"}})
    fireEvent.click(screen.getByText("Check"))
}

describe("MoveContainerForm", () => {
    it("keeps Check disabled until a container is picked", () => {
        renderForm()

        expect(screen.getByText("Check").closest("button")).toBeDisabled()
    })

    it("keeps Start disabled until network consumers are acknowledged", async () => {
        planTransfer.mockResolvedValue({
            sidecars: [{containerName: "app-redis", imageName: "redis:7"}],
            consumers: [{containerName: "web", networkName: "shared"}],
            unsupportedMounts: [],
        })
        renderForm()

        fillAndCheck()

        await waitFor(() => expect(screen.getByText(/will lose access/)).toBeInTheDocument())
        expect(screen.getByText("web")).toBeInTheDocument()
        expect(screen.getByText("Start transfer").closest("button")).toBeDisabled()

        fireEvent.click(screen.getByText("I understand"))

        await waitFor(() => expect(screen.getByText("Start transfer").closest("button")).toBeEnabled())
    })

    it("blocks Start when the plan has unsupported host mounts", async () => {
        planTransfer.mockResolvedValue({sidecars: [], consumers: [], unsupportedMounts: ["/etc/host"]})
        renderForm()

        fillAndCheck()

        await waitFor(() => expect(screen.getByText("/etc/host")).toBeInTheDocument())
        expect(screen.getByText("Start transfer").closest("button")).toBeDisabled()
    })

    it("starts the transfer with every sidecar ticked by default", async () => {
        planTransfer.mockResolvedValue({
            sidecars: [{containerName: "app-redis", imageName: "redis:7"}],
            consumers: [],
            unsupportedMounts: [],
        })
        renderForm()

        fillAndCheck()
        await waitFor(() => expect(screen.getByText("Start transfer").closest("button")).toBeEnabled())
        fireEvent.click(screen.getByText("Start transfer"))

        await waitFor(() => expect(startTransfer).toHaveBeenCalledWith({
            containerName: "app",
            sidecarNames: ["app-redis"],
            address: "10.0.0.2:53890",
            token: "tok",
            isConsumersAcknowledged: false,
        }))
    })
})
