import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {ServicePattern, type DockerContainer} from "@/app/api/velez"
import {IsStatefullModeEnabled} from "@/processes/queries/control_plane.ts"
import {useRegisterPreviewQuery} from "@/processes/queries/containers.ts"
import CreateServiceDialog from "@/dialogs/CreateServiceDialog/CreateServiceDialog.tsx"
import {ServiceScreen} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"

vi.mock("@/processes/queries/control_plane.ts", () => ({IsStatefullModeEnabled: vi.fn()}))
vi.mock("@/processes/queries/containers.ts", () => ({useRegisterPreviewQuery: vi.fn()}))
vi.mock("@/dialogs/CreateServiceDialog/screens/GenericAdoptScreen/GenericAdoptScreen.tsx", () => ({
    default: () => <span>generic adopt screen</span>,
}))
vi.mock("@/dialogs/CreateServiceDialog/screens/PostgresAdoptScreen/PostgresAdoptScreen.tsx", () => ({
    default: () => <span>postgres adopt screen</span>,
}))

vi.mock("@/dialogs/CreateServiceDialog/screens/GenericScreen/GenericScreen.tsx", () => ({
    default: () => <span>generic screen</span>,
}))
vi.mock("@/dialogs/CreateServiceDialog/screens/PostgresScreen/PostgresScreen.tsx", () => ({
    default: () => <span>postgres screen</span>,
}))
vi.mock("@/dialogs/CreateServiceDialog/screens/RegistryScreen/RegistryScreen.tsx", () => ({
    default: () => <span>registry screen</span>,
}))
vi.mock("@/dialogs/CreateServiceDialog/screens/RunnerScreen/RunnerScreen.tsx", () => ({
    default: () => <span>runner screen</span>,
}))

const ROOT: DockerContainer = {
    id: "root",
    name: "vpn",
    imageName: "tailscale:1",
    suggestedPattern: ServicePattern.SERVICE_PATTERN_POSTGRES,
}
const MEMBER: DockerContainer = {id: "app", name: "app", imageName: "nginx:1"}

function renderDialog(initialScreen?: ServiceScreen) {
    render(<CreateServiceDialog initialScreen={initialScreen}/>)
}

function renderAdoptDialog(networkMembers: DockerContainer[] = [MEMBER]) {
    vi.mocked(IsStatefullModeEnabled).mockReturnValue(false)
    vi.mocked(useRegisterPreviewQuery).mockReturnValue({
        data: {containers: []},
        isLoading: false,
        isError: false,
        refetch: vi.fn(),
    } as Partial<ReturnType<typeof useRegisterPreviewQuery>> as ReturnType<typeof useRegisterPreviewQuery>)

    render(<CreateServiceDialog container={ROOT} networkMembers={networkMembers}/>)
}

function itemButton(name: string): HTMLElement {
    return screen.getByRole("button", {name: new RegExp(`^${name}`)})
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("CreateServiceDialog", () => {
    it("opens on the picker by default", () => {
        renderDialog()

        expect(screen.getByText("Create service")).toBeInTheDocument()
        expect(screen.getByText("PostgreSQL")).toBeInTheDocument()
        expect(screen.queryByText("Back")).not.toBeInTheDocument()
    })

    it("skips the picker and hides Back when opened on a specific screen", () => {
        renderDialog("postgres")

        expect(screen.getByText("postgres screen")).toBeInTheDocument()
        expect(screen.getByText("Create database")).toBeInTheDocument()
        expect(screen.queryByText("PostgreSQL")).not.toBeInTheDocument()
        expect(screen.queryByText("Back")).not.toBeInTheDocument()
    })

    it("shows the picked screen with a Back button that returns to the picker", () => {
        renderDialog()

        fireEvent.click(screen.getByText("Container registry"))

        expect(screen.getByText("registry screen")).toBeInTheDocument()
        expect(screen.getByText("Create registry")).toBeInTheDocument()

        fireEvent.click(screen.getByText("Back"))

        expect(screen.getByText("Create service")).toBeInTheDocument()
        expect(screen.queryByText("registry screen")).not.toBeInTheDocument()
    })

    it("lists the root and every member as accordion items when onboarding a network group", () => {
        renderAdoptDialog()

        expect(itemButton("vpn")).toHaveAttribute("aria-expanded", "false")
        expect(itemButton("app")).toHaveAttribute("aria-expanded", "false")
        expect(itemButton("vpn")).toHaveTextContent("Service")
        expect(itemButton("vpn")).toHaveTextContent("Network root")
        expect(itemButton("app")).not.toHaveTextContent("Network root")
    })

    it("lists a single item without role buttons when the container has no members", () => {
        renderAdoptDialog([])

        fireEvent.click(itemButton("vpn"))

        expect(screen.queryByRole("button", {name: "Make service"})).not.toBeInTheDocument()
        expect(screen.queryByRole("button", {name: "Make network root"})).not.toBeInTheDocument()
    })

    it("moves the service badge when Make service is clicked, without a confirmation", () => {
        renderAdoptDialog()
        fireEvent.click(itemButton("app"))

        fireEvent.click(screen.getByRole("button", {name: "Make service"}))

        expect(itemButton("app")).toHaveTextContent("Service")
        expect(itemButton("vpn")).not.toHaveTextContent("Service")
        expect(itemButton("vpn")).toHaveTextContent("Network root")
    })

    it("forces the generic screen when another container becomes the service", () => {
        renderAdoptDialog()
        fireEvent.click(screen.getByRole("button", {name: /PostgreSQL/}))
        expect(screen.getByText("postgres adopt screen")).toBeInTheDocument()

        fireEvent.click(itemButton("app"))
        fireEvent.click(screen.getByRole("button", {name: "Make service"}))

        expect(screen.getByText("generic adopt screen")).toBeInTheDocument()
        expect(screen.queryByText("postgres adopt screen")).not.toBeInTheDocument()
        expect(screen.getByText(/only the Generic pattern is available/)).toBeInTheDocument()
    })

    it("asks for confirmation before switching the network root and changes nothing on cancel", () => {
        renderAdoptDialog()
        fireEvent.click(itemButton("app"))

        fireEvent.click(screen.getByRole("button", {name: "Make network root"}))

        expect(screen.getByText("Make network root?")).toBeInTheDocument()
        expect(screen.getByText(/briefly unavailable/)).toBeInTheDocument()

        fireEvent.click(screen.getByRole("button", {name: "Cancel"}))

        expect(screen.queryByText("Make network root?")).not.toBeInTheDocument()
        expect(itemButton("vpn")).toHaveTextContent("Network root")
        expect(itemButton("app")).not.toHaveTextContent("Network root")
    })

    it("switches the network root and the service along with it once confirmed", () => {
        renderAdoptDialog()
        fireEvent.click(itemButton("app"))
        fireEvent.click(screen.getByRole("button", {name: "Make network root"}))

        fireEvent.click(screen.getByRole("button", {name: "Make network root"}))

        expect(screen.queryByText("Make network root?")).not.toBeInTheDocument()
        expect(itemButton("app")).toHaveTextContent("Network root")
        expect(itemButton("app")).toHaveTextContent("Service")
        expect(itemButton("vpn")).not.toHaveTextContent("Network root")
    })
})
