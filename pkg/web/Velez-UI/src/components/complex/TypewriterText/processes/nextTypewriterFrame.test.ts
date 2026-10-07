import {describe, expect, it} from "vitest"

import {frameDelayMs, nextTypewriterFrame} from "@/components/complex/TypewriterText/processes/nextTypewriterFrame.ts"

describe("nextTypewriterFrame", () => {
    it("does stay put when shown equals the target", () => {
        expect(nextTypewriterFrame("Queued", "Queued")).toBe("Queued")
    })

    it("does type one more character when shown is a prefix of the target", () => {
        expect(nextTypewriterFrame("Que", "Queued")).toBe("Queu")
        expect(nextTypewriterFrame("", "Hi")).toBe("H")
    })

    it("does delete one character when shown is not a prefix of the target", () => {
        expect(nextTypewriterFrame("Queued", "Applied")).toBe("Queue")
    })

    it("does delete when the target is a strict prefix of shown", () => {
        expect(nextTypewriterFrame("Queued", "Que")).toBe("Queue")
    })

    it("does reach the target by deleting then typing from any starting point", () => {
        let shown = "Pulling image"
        for (let i = 0; i < 100 && shown !== "Applied"; i++) {
            shown = nextTypewriterFrame(shown, "Applied")
        }

        expect(shown).toBe("Applied")
    })
})

describe("frameDelayMs", () => {
    it("does use the faster delay when deleting and the slower when typing", () => {
        expect(frameDelayMs("ab", "a")).toBe(20)
        expect(frameDelayMs("a", "ab")).toBe(30)
    })
})
