import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import BindMountLinks from "@/dialogs/CreateServiceDialog/components/BindMountLinks/BindMountLinks.tsx"

const LINKS = [
    {
        source: "/srv/pg",
        destination: "/var/lib/postgresql/data",
        volumeName: "db_var_lib_postgresql_data",
        isKeptAsBind: false,
    },
    {source: "/srv/conf", destination: "/etc/app", volumeName: "db_etc_app", isKeptAsBind: false},
]

const SOCKET_LINK = {
    source: "/var/run/docker.sock",
    destination: "/docker.sock",
    volumeName: "db_docker.sock",
    isKeptAsBind: true,
}

function renderLinks(links = LINKS) {
    const onVolumeNameChange = vi.fn()
    const onKeptChange = vi.fn()
    render(<BindMountLinks links={links} onVolumeNameChange={onVolumeNameChange} onKeptChange={onKeptChange}/>)
    return {onVolumeNameChange, onKeptChange}
}

describe("BindMountLinks", () => {
    it("shows every bind mount source and destination with its prefilled volume name", () => {
        renderLinks()

        expect(screen.getByText("/srv/pg")).toBeInTheDocument()
        expect(screen.getByText("/var/lib/postgresql/data")).toBeInTheDocument()
        expect(screen.getByText("/srv/conf")).toBeInTheDocument()
        expect(screen.getByText("/etc/app")).toBeInTheDocument()
        expect(screen.getAllByRole("textbox").map((input) => (input as HTMLInputElement).value))
            .toEqual(["db_var_lib_postgresql_data", "db_etc_app"])
    })

    it("explains linking, keeping and system paths", () => {
        renderLinks()

        expect(screen.getByText(/nothing is copied/)).toBeInTheDocument()
        expect(screen.getByText(/System paths such as \/dev and \/var\/run are kept/)).toBeInTheDocument()
    })

    it("reports the edited volume name against the mount target", () => {
        const {onVolumeNameChange} = renderLinks()

        fireEvent.change(screen.getAllByRole("textbox")[1], {target: {value: "conf"}})

        expect(onVolumeNameChange).toHaveBeenCalledWith("/etc/app", "conf")
    })

    it("hides the volume name of a mount that is kept as a bind mount", () => {
        renderLinks([SOCKET_LINK, LINKS[0]])

        expect(screen.getAllByRole("textbox")).toHaveLength(1)
        const [keptCheckbox, plainCheckbox] = screen.getAllByRole("checkbox", {name: "Keep as bind mount"})

        expect(keptCheckbox).toBeChecked()
        expect(plainCheckbox).not.toBeChecked()
    })

    it("reports the keep toggle against the mount target", () => {
        const {onKeptChange} = renderLinks()

        fireEvent.click(screen.getAllByRole("checkbox", {name: "Keep as bind mount"})[1])

        expect(onKeptChange).toHaveBeenCalledWith("/etc/app", true)
    })
})
