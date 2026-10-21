import {beforeEach, describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"

import RootChangeWarning from "@/dialogs/CreateServiceDialog/components/RootChangeWarning/RootChangeWarning.tsx"
import {IsStatefullModeEnabled} from "@/processes/queries/control_plane.ts"

vi.mock("@/processes/queries/control_plane.ts", () => ({IsStatefullModeEnabled: vi.fn()}))

describe("RootChangeWarning", () => {
    beforeEach(() => {
        vi.mocked(IsStatefullModeEnabled).mockReturnValue(false)
    })

    it("shows the lead and the switch message when visible", () => {
        render(<RootChangeWarning isVisible oldRootName="app" newRootName="vpn"/>)

        const status = screen.getByRole("status")
        expect(status).toHaveTextContent("Network root changes")
        expect(status).toHaveTextContent("vpn becomes the network root and app rejoins its network as a sidecar.")
    })

    it("is hidden from the accessibility tree when not visible", () => {
        render(<RootChangeWarning isVisible={false} oldRootName="app" newRootName="vpn"/>)

        expect(screen.queryByRole("status")).not.toBeInTheDocument()
    })

    it("adds the cluster mode sentence in cluster mode", () => {
        vi.mocked(IsStatefullModeEnabled).mockReturnValue(true)

        render(<RootChangeWarning isVisible oldRootName="app" newRootName="vpn"/>)

        expect(screen.getByRole("status")).toHaveTextContent("In cluster mode the containers are recreated too")
    })
})
