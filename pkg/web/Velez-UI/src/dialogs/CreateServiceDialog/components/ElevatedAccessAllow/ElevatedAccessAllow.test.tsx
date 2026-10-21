import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import ElevatedAccessAllow from "@/dialogs/CreateServiceDialog/components/ElevatedAccessAllow/ElevatedAccessAllow.tsx"

function renderAllow(isChecked: boolean) {
    const onChange = vi.fn()
    render(<ElevatedAccessAllow repository="tailscale/tailscale" isChecked={isChecked} onChange={onChange}/>)
    return {onChange}
}

describe("ElevatedAccessAllow", () => {
    it("names the image repository and explains why allowing is needed", () => {
        renderAllow(false)

        expect(screen.getByRole("switch", {name: "Allow elevated access for tailscale/tailscale"}))
            .toHaveAttribute("aria-checked", "false")
        expect(screen.getByText(/This image is not whitelisted/)).toBeInTheDocument()
    })

    it("reports the flipped value when the toggle is clicked", () => {
        const {onChange} = renderAllow(false)

        fireEvent.click(screen.getByRole("switch"))

        expect(onChange).toHaveBeenCalledWith(true)
    })
})
