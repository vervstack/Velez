import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import RegistryInstanceCreateDialog from "@/dialogs/RegistryInstanceCreateDialog/RegistryInstanceCreateDialog.tsx"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {ListEnvironmentsQuery} from "@/processes/queries/control_plane.ts"
import {useListServicesQuery} from "@/processes/queries/services.ts"
import {CreateRegistryInstanceMutation} from "@/processes/queries/registry_instances.ts"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/processes/queries/control_plane.ts", () => ({ListEnvironmentsQuery: vi.fn()}))
vi.mock("@/processes/queries/services.ts", () => ({useListServicesQuery: vi.fn()}))
vi.mock("@/processes/queries/registry_instances.ts", () => ({CreateRegistryInstanceMutation: vi.fn()}))

function renderDialog() {
    const CloseDialog = vi.fn()
    vi.mocked(useDialog).mockReturnValue(
        {CloseDialog} as Partial<ReturnType<typeof useDialog>> as ReturnType<typeof useDialog>
    )
    vi.mocked(ListEnvironmentsQuery).mockReturnValue(
        {data: {environments: []}} as Partial<ReturnType<typeof ListEnvironmentsQuery>> as
            ReturnType<typeof ListEnvironmentsQuery>
    )
    vi.mocked(useListServicesQuery).mockReturnValue(
        {data: {services: []}} as Partial<ReturnType<typeof useListServicesQuery>> as
            ReturnType<typeof useListServicesQuery>
    )
    const mutate = vi.fn()
    vi.mocked(CreateRegistryInstanceMutation).mockReturnValue(
        {mutate, isPending: false} as Partial<ReturnType<typeof CreateRegistryInstanceMutation>> as
            ReturnType<typeof CreateRegistryInstanceMutation>
    )

    render(<RegistryInstanceCreateDialog/>)
    return {mutate, CloseDialog}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("RegistryInstanceCreateDialog", () => {
    it("disables Create until a name is entered", () => {
        renderDialog()

        expect(screen.getByText("Create")).toBeDisabled()

        fireEvent.change(screen.getByRole("textbox"), {target: {value: "my-registry"}})

        expect(screen.getByText("Create")).not.toBeDisabled()
    })

    it("submits a create request built from the form fields", () => {
        const {mutate} = renderDialog()

        fireEvent.change(screen.getByRole("textbox"), {target: {value: "my-registry"}})
        fireEvent.click(screen.getByText("Create"))

        expect(mutate).toHaveBeenCalledWith(
            expect.objectContaining({name: "my-registry", box: "small"}),
            expect.anything()
        )
    })
})
