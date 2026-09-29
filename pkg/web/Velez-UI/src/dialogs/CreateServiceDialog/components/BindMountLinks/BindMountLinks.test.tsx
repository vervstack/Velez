import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import BindMountLinks from "@/dialogs/CreateServiceDialog/components/BindMountLinks/BindMountLinks.tsx"

const LINKS = [
    {source: "/srv/pg", destination: "/var/lib/postgresql/data", volumeName: "db_var_lib_postgresql_data"},
    {source: "/srv/conf", destination: "/etc/app", volumeName: "db_etc_app"},
]

describe("BindMountLinks", () => {
    it("shows every bind mount as source to target with its prefilled volume name", () => {
        render(<BindMountLinks links={LINKS} onVolumeNameChange={vi.fn()}/>)

        expect(screen.getByText("/srv/pg → /var/lib/postgresql/data")).toBeInTheDocument()
        expect(screen.getByText("/srv/conf → /etc/app")).toBeInTheDocument()
        expect(screen.getAllByRole("textbox").map((input) => (input as HTMLInputElement).value))
            .toEqual(["db_var_lib_postgresql_data", "db_etc_app"])
    })

    it("explains what linking does", () => {
        render(<BindMountLinks links={LINKS} onVolumeNameChange={vi.fn()}/>)

        expect(screen.getByText(/nothing is copied/)).toBeInTheDocument()
    })

    it("reports the edited volume name against the mount target", () => {
        const onVolumeNameChange = vi.fn()
        render(<BindMountLinks links={LINKS} onVolumeNameChange={onVolumeNameChange}/>)

        fireEvent.change(screen.getAllByRole("textbox")[1], {target: {value: "conf"}})

        expect(onVolumeNameChange).toHaveBeenCalledWith("/etc/app", "conf")
    })
})
