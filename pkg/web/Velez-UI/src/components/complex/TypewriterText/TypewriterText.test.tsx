import {act, render, screen} from "@testing-library/react"
import {afterEach, beforeEach, describe, expect, it, vi} from "vitest"

import TypewriterText from "@/components/complex/TypewriterText/TypewriterText.tsx"

const STEP_MS = 10

// Each timer is scheduled by an effect that only runs once act() flushes, so time moves in small acts.
function advance(ms: number) {
    for (let elapsed = 0; elapsed < ms; elapsed += STEP_MS) {
        act(() => {
            vi.advanceTimersByTime(STEP_MS)
        })
    }
}

function stubReducedMotion(isReduced: boolean) {
    vi.stubGlobal("matchMedia", vi.fn().mockReturnValue({matches: isReduced}))
}

describe("TypewriterText", () => {
    beforeEach(() => {
        vi.useFakeTimers()
        stubReducedMotion(false)
    })

    afterEach(() => {
        vi.useRealTimers()
        vi.unstubAllGlobals()
    })

    it("does type the text one character at a time", () => {
        render(<TypewriterText text="Hey"/>)

        expect(screen.queryByText("H")).not.toBeInTheDocument()
        advance(30)
        expect(screen.getByText("H")).toBeInTheDocument()
        advance(30)
        expect(screen.getByText("He")).toBeInTheDocument()
        advance(30)
        expect(screen.getByText("Hey")).toBeInTheDocument()
    })

    it("does delete the shown text then type the new text when the text changes", () => {
        const {rerender} = render(<TypewriterText text="Hey"/>)
        advance(90)

        rerender(<TypewriterText text="Yo"/>)
        advance(20)
        expect(screen.getByText("He")).toBeInTheDocument()
        advance(40)
        expect(screen.queryByText(/^H/)).not.toBeInTheDocument()
        advance(30)
        expect(screen.getByText("Y")).toBeInTheDocument()
        advance(30)
        expect(screen.getByText("Yo")).toBeInTheDocument()
    })

    it("does retarget toward the latest text when it changes mid-animation", () => {
        const {rerender} = render(<TypewriterText text="Alpha"/>)
        advance(60)
        expect(screen.getByText("Al")).toBeInTheDocument()

        rerender(<TypewriterText text="Beta"/>)
        rerender(<TypewriterText text="Gamma"/>)
        advance(2000)

        expect(screen.getByText("Gamma")).toBeInTheDocument()
    })

    it("does show the text instantly when reduced motion is preferred", () => {
        stubReducedMotion(true)

        render(<TypewriterText text="Instant"/>)

        expect(screen.getByText("Instant")).toBeInTheDocument()
    })
})
