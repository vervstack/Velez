import {afterEach, beforeEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"
import {MemoryRouter} from "react-router-dom"

import {EnablePluginResponse, VervPluginState, VervPluginType} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import Button from "@/components/base/Button.tsx"
import RegistryPluginForm from "@/dialogs/PluginManageDialog/plugins/RegistryPluginForm.tsx"
import {VervPlugin} from "@/model/services/VervPlugins.tsx"
import {controlPlaneService} from "@/processes/api/control_plane.ts"
import {serviceService} from "@/processes/api/service.ts"

vi.mock("@/processes/api/control_plane.ts", () => ({
    controlPlaneService: {
        enableRegistry: vi.fn(),
    },
}))

vi.mock("@/processes/api/service.ts", () => ({
    serviceService: {
        restartService: vi.fn().mockResolvedValue(undefined),
    },
}))

interface StartProp {
    start(): Promise<EnablePluginResponse>
}

vi.mock("@/dialogs/PluginManageDialog/plugins/screens/TaskProgressScreen.tsx", () => ({
    default: ({start}: StartProp) => {
        function handleStartTask() {
            start()
        }

        return <Button onClick={handleStartTask}>start-task</Button>
    },
}))

function renderForm(state: VervPluginState) {
    const plugin = new VervPlugin(VervPluginType.registry, "registry")
    plugin.state = state

    return render(
        <MemoryRouter>
            <RegistryPluginForm {...plugin}/>
        </MemoryRouter>
    )
}

beforeEach(() => {
    useDialog.setState({children: null, IsClickOffClosesDialog: true})
    vi.mocked(controlPlaneService.enableRegistry).mockResolvedValue({
        entityId: "entity-1",
        action: "enable_registry",
    })
})

afterEach(() => {
    vi.restoreAllMocks()
})

describe("RegistryPluginForm", () => {
    it("enables the registry with the entered port and username once Enable is clicked", () => {
        renderForm(VervPluginState.disabled)

        const [portInput, usernameInput] = screen.getAllByRole("textbox")
        fireEvent.change(portInput, {target: {value: "5001"}})
        fireEvent.change(usernameInput, {target: {value: "alice"}})
        fireEvent.click(screen.getByText("Enable"))
        fireEvent.click(screen.getByText("start-task"))

        expect(controlPlaneService.enableRegistry).toHaveBeenCalledWith({exposeToPort: 5001, username: "alice"})
    })

    it("omits port and username when left blank", () => {
        renderForm(VervPluginState.disabled)

        fireEvent.click(screen.getByText("Enable"))
        fireEvent.click(screen.getByText("start-task"))

        expect(controlPlaneService.enableRegistry).toHaveBeenCalledWith({
            exposeToPort: undefined,
            username: undefined,
        })
    })

    it("offers a restart when the registry service is dead", () => {
        renderForm(VervPluginState.dead)

        fireEvent.click(screen.getByText("Restart"))

        expect(serviceService.restartService).toHaveBeenCalledWith("registry")
    })
})
