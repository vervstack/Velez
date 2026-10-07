import {describe, expect, it} from "vitest"

import {
    layoutSidecarGroup,
    SIDECAR_CHIP_H,
    SIDECAR_CHIP_W,
    SIDECAR_LABEL_MAX_CHARS,
    sidecarDotTone,
    truncateSidecarLabel,
} from "@/processes/mappings/serviceGraphLayout"

const base = {centerX: 490, centerY: 350, ringRadius: 62, labelBottomY: 440}

describe("layoutSidecarGroup", () => {
    it("centers a single chip under the service label", () => {
        const layout = layoutSidecarGroup({...base, sidecarCount: 1})

        expect(layout.chips).toHaveLength(1)
        expect(layout.chips[0].x + SIDECAR_CHIP_W / 2).toBe(base.centerX)
        expect(layout.chips[0].y).toBeGreaterThan(base.labelBottomY)
    })

    it("lays three chips on one row, symmetric around the center", () => {
        const layout = layoutSidecarGroup({...base, sidecarCount: 3})

        const rows = new Set(layout.chips.map(chip => chip.y))
        const lefts = layout.chips.map(chip => chip.x)
        const rightEdge = lefts[2] + SIDECAR_CHIP_W

        expect(rows.size).toBe(1)
        expect(base.centerX - lefts[0]).toBe(rightEdge - base.centerX)
    })

    it("wraps to a second row after three chips", () => {
        const layout = layoutSidecarGroup({...base, sidecarCount: 4})

        const firstRowY = layout.chips[0].y
        expect(layout.chips[3].y).toBeGreaterThan(firstRowY + SIDECAR_CHIP_H)
        expect(layout.chips.map(chip => chip.index)).toEqual([0, 1, 2, 3])
    })

    it("encloses the ring, every chip and the caption in the outline", () => {
        const layout = layoutSidecarGroup({...base, sidecarCount: 5})
        const {outline} = layout

        expect(outline.y).toBeLessThan(base.centerY - base.ringRadius)
        expect(outline.x).toBeLessThan(base.centerX - base.ringRadius)
        expect(layout.captionY).toBeGreaterThan(outline.y)
        for (const chip of layout.chips) {
            expect(chip.x).toBeGreaterThan(outline.x)
            expect(chip.x + chip.width).toBeLessThan(outline.x + outline.width)
            expect(chip.y + chip.height).toBeLessThan(outline.y + outline.height)
        }
    })

    it("draws one connector drop per chip plus a bus per row and a spine", () => {
        const layout = layoutSidecarGroup({...base, sidecarCount: 4})

        expect(layout.connectors).toHaveLength(4 + 2 + 1)
    })

    it("keeps the outline centered on the service node", () => {
        const layout = layoutSidecarGroup({...base, sidecarCount: 2})

        expect(layout.outline.x + layout.outline.width / 2).toBe(base.centerX)
    })
})

describe("truncateSidecarLabel", () => {
    it("keeps a name that fits", () => {
        expect(truncateSidecarLabel("s3_artel_web_ui")).toBe("s3_artel_web_ui")
    })

    it("cuts a long name to the limit with an ellipsis", () => {
        const result = truncateSidecarLabel("a".repeat(40))

        expect(result).toHaveLength(SIDECAR_LABEL_MAX_CHARS)
        expect(result.endsWith("…")).toBe(true)
    })
})

describe("sidecarDotTone", () => {
    it("maps dot statuses to a tone", () => {
        expect(sidecarDotTone("running")).toBe("ok")
        expect(sidecarDotTone("pending")).toBe("warn")
        expect(sidecarDotTone("error")).toBe("bad")
        expect(sidecarDotTone("offline")).toBe("idle")
    })
})
