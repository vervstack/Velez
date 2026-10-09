import {render, screen} from "@testing-library/react"
import {describe, expect, it} from "vitest"

import SlidingLabel from "@/components/complex/SlidingLabel/SlidingLabel.tsx"

function renderLabel(isSwapped: boolean) {
    render(<SlidingLabel label="BuildKit: on" swapLabel="Disable BuildKit" isSwapped={isSwapped}/>)
}

describe("SlidingLabel", () => {
    it("exposes only the resting label to assistive tech when not swapped", () => {
        renderLabel(false)

        expect(screen.getByText("BuildKit: on").getAttribute("aria-hidden")).toBe("false")
        expect(screen.getByText("Disable BuildKit").getAttribute("aria-hidden")).toBe("true")
    })

    it("exposes only the swap label to assistive tech when swapped", () => {
        renderLabel(true)

        expect(screen.getByText("BuildKit: on").getAttribute("aria-hidden")).toBe("true")
        expect(screen.getByText("Disable BuildKit").getAttribute("aria-hidden")).toBe("false")
    })
})
