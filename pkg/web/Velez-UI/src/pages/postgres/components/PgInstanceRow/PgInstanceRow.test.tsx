import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"
import {MemoryRouter} from "react-router-dom"

import type {PgInstance} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import PgInstanceRow from "@/pages/postgres/components/PgInstanceRow/PgInstanceRow.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/pages/postgres/components/PgInstanceCredentials/PgInstanceCredentials.tsx", () => ({default: () => null}))

function renderRow(instance: PgInstance) {
    const OpenDialog = vi.fn()
    vi.mocked(useDialog).mockReturnValue(
        {OpenDialog} as Partial<ReturnType<typeof useDialog>> as ReturnType<typeof useDialog>
    )

    render(
        <MemoryRouter>
            <PgInstanceRow instance={instance}/>
        </MemoryRouter>
    )
    return {OpenDialog}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("PgInstanceRow", () => {
    it("shows the display name and keeps the instance name in the link", () => {
        renderRow({name: "pgaas_orders", displayName: "orders"})

        expect(screen.getByRole("link", {name: "orders"}))
            .toHaveAttribute("href", expect.stringContaining("pgaas_orders"))
    })

    it("falls back to the instance name when there is no display name", () => {
        renderRow({name: "pgaas_orders", displayName: ""})

        expect(screen.getByRole("link", {name: "pgaas_orders"})).toBeInTheDocument()
    })

    it("opens the drop dialog with the instance name when Drop is clicked", () => {
        const {OpenDialog} = renderRow({name: "pgaas_orders", displayName: "orders"})

        fireEvent.click(screen.getByRole("button", {name: "Drop"}))

        expect(OpenDialog.mock.calls[0][0].props.name).toBe("pgaas_orders")
    })
})
