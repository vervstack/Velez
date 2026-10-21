import {describe, expect, it} from "vitest"
import {render, screen} from "@testing-library/react"

import type {ElevatedAccess} from "@/app/api/velez"
import ElevatedAccessChips from "@/dialogs/CreateServiceDialog/components/ElevatedAccessChips/ElevatedAccessChips.tsx"

function renderChips(access: ElevatedAccess | undefined, isRecreating = true) {
    return render(<ElevatedAccessChips access={access} isRecreating={isRecreating}/>)
}

describe("ElevatedAccessChips", () => {
    it("renders nothing when the container is not recreated", () => {
        const {container} = renderChips({isPrivileged: true}, false)

        expect(container).toBeEmptyDOMElement()
    })

    it("shows one chip per item and says they are carried over", () => {
        renderChips({
            isPrivileged: true,
            devices: [{hostPath: "/dev/net/tun"}],
            capAdd: ["NET_ADMIN", "SYS_TIME"],
            isHostPid: true,
            isHostIpc: true,
            isHostUserns: true,
        })

        for (const label of [
            "Privileged", "Device /dev/net/tun", "Cap NET_ADMIN", "Cap SYS_TIME",
            "Host PID", "Host IPC", "Host user namespace",
        ]) {
            expect(screen.getByText(label)).toBeInTheDocument()
        }
        expect(screen.getByText("Carried over to the new container.")).toBeInTheDocument()
    })

    it("shows a single neutral chip when the container has no elevated access", () => {
        renderChips(undefined)

        expect(screen.getByText("No elevated access")).toBeInTheDocument()
        expect(screen.queryByText("Carried over to the new container.")).not.toBeInTheDocument()
    })
})
