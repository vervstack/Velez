import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"
import {useNavigate} from "react-router-dom"

import {SmerdStatus} from "@/app/api/velez"
import ServiceHeroImages from "@/widgets/service/ServiceHeroImages/ServiceHeroImages.tsx"

vi.mock("react-router-dom", () => ({useNavigate: vi.fn()}))

function renderImages(props: Partial<Parameters<typeof ServiceHeroImages>[0]>) {
    const navigate = vi.fn()
    vi.mocked(useNavigate).mockReturnValue(
        navigate as Partial<ReturnType<typeof useNavigate>> as ReturnType<typeof useNavigate>
    )

    const {container} = render(<ServiceHeroImages serviceName="zpotify" sidecars={[]} {...props}/>)
    return {navigate, container}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("ServiceHeroImages", () => {
    it("opens the main container page when the service image is clicked", () => {
        const {navigate} = renderImages({imageName: "redsockruf/zpotify", containerId: "main-1"})

        fireEvent.click(screen.getByRole("button", {name: /zpotify/}))

        expect(navigate).toHaveBeenCalledWith("/container/main-1")
    })

    it("draws each sidecar and opens its own container page", () => {
        const {navigate} = renderImages({
            imageName: "redsockruf/zpotify",
            containerId: "main-1",
            sidecars: [
                {containerId: "side-1", name: "vpn", imageName: "tailscale/tailscale", status: SmerdStatus.running},
                {containerId: "side-2", name: "proxy", imageName: "nginx", status: SmerdStatus.exited},
            ],
        })

        fireEvent.click(screen.getByRole("button", {name: /proxy/}))

        expect(screen.getByRole("button", {name: /vpn/})).toBeTruthy()
        expect(navigate).toHaveBeenCalledWith("/container/side-2")
    })

    it("renders nothing without a main container and without sidecars", () => {
        const {container} = renderImages({})

        expect(container.firstChild).toBeNull()
    })
})
