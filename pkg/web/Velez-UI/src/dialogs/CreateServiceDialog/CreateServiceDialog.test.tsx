import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import CreateServiceDialog from "@/dialogs/CreateServiceDialog/CreateServiceDialog.tsx"
import {ServiceScreen} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"

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

function renderDialog(initialScreen?: ServiceScreen) {
    render(<CreateServiceDialog initialScreen={initialScreen}/>)
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
})
