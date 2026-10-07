import type {DotStatus} from "@/components/base/StatusDot.tsx"

export interface Rect {
    x: number
    y: number
    width: number
    height: number
}

export interface SidecarChipLayout extends Rect {
    index: number
}

export interface SidecarGroupInput {
    centerX: number
    centerY: number
    ringRadius: number
    labelBottomY: number
    sidecarCount: number
}

export interface SidecarGroupLayout {
    outline: Rect
    captionX: number
    captionY: number
    chips: SidecarChipLayout[]
    connectors: string[]
}

export type SidecarDotTone = "ok" | "warn" | "bad" | "idle"

export const SIDECAR_CHIP_W = 140
export const SIDECAR_CHIP_H = 26
export const SIDECAR_LABEL_MAX_CHARS = 18

const CHIP_GAP = 10
const CHIPS_PER_ROW = 3
const STEM_LENGTH = 10
const BUS_TO_CHIP = 12
const ROW_GAP = 14
const ROW_PITCH = BUS_TO_CHIP + SIDECAR_CHIP_H + ROW_GAP
const GROUP_PAD_X = 22
const GROUP_PAD_TOP = 26
const GROUP_PAD_BOTTOM = 14
const CAPTION_OFFSET = 12

const DOT_TONES: Record<DotStatus, SidecarDotTone> = {
    running: "ok",
    healthy: "ok",
    online: "ok",
    enabled: "ok",
    degraded: "warn",
    pending: "warn",
    creating: "warn",
    error: "bad",
    stopped: "idle",
    offline: "idle",
    disabled: "idle",
}

export function sidecarDotTone(status: DotStatus): SidecarDotTone {
    return DOT_TONES[status]
}

export function truncateSidecarLabel(name: string): string {
    if (name.length <= SIDECAR_LABEL_MAX_CHARS) return name
    return name.slice(0, SIDECAR_LABEL_MAX_CHARS - 1) + "…"
}

function rowWidth(chipsInRow: number): number {
    return chipsInRow * SIDECAR_CHIP_W + (chipsInRow - 1) * CHIP_GAP
}

function chipsInRow(row: number, total: number): number {
    return Math.min(CHIPS_PER_ROW, total - row * CHIPS_PER_ROW)
}

function chipRowCount(total: number): number {
    return Math.ceil(total / CHIPS_PER_ROW)
}

export function layoutSidecarGroup(input: SidecarGroupInput): SidecarGroupLayout {
    const {centerX, centerY, ringRadius, labelBottomY, sidecarCount} = input
    const rowCount = chipRowCount(sidecarCount)
    const firstBusY = labelBottomY + STEM_LENGTH

    const chips: SidecarChipLayout[] = []
    const connectors: string[] = []

    for (let row = 0; row < rowCount; row++) {
        const inRow = chipsInRow(row, sidecarCount)
        const busY = firstBusY + row * ROW_PITCH
        const chipTop = busY + BUS_TO_CHIP
        const left = centerX - rowWidth(inRow) / 2

        for (let col = 0; col < inRow; col++) {
            const chipLeft = left + col * (SIDECAR_CHIP_W + CHIP_GAP)
            const chipCenterX = chipLeft + SIDECAR_CHIP_W / 2
            const chip: SidecarChipLayout = {
                x: chipLeft,
                y: chipTop,
                width: SIDECAR_CHIP_W,
                height: SIDECAR_CHIP_H,
                index: row * CHIPS_PER_ROW + col,
            }
            chips.push(chip)
            connectors.push(`M ${chipCenterX} ${busY} V ${chipTop}`)
        }

        const busLeft = left + SIDECAR_CHIP_W / 2
        const busRight = left + rowWidth(inRow) - SIDECAR_CHIP_W / 2
        connectors.push(`M ${busLeft} ${busY} H ${busRight}`)
    }

    const lastBusY = firstBusY + (rowCount - 1) * ROW_PITCH
    connectors.push(`M ${centerX} ${labelBottomY} V ${lastBusY}`)

    const widestRow = chipsInRow(0, sidecarCount)
    const contentWidth = Math.max(ringRadius * 2, rowWidth(widestRow))
    const width = contentWidth + GROUP_PAD_X * 2
    const top = centerY - ringRadius - GROUP_PAD_TOP
    const lastChipBottom = lastBusY + BUS_TO_CHIP + SIDECAR_CHIP_H

    const outline: Rect = {
        x: centerX - width / 2,
        y: top,
        width,
        height: lastChipBottom + GROUP_PAD_BOTTOM - top,
    }

    return {outline, captionX: centerX, captionY: top + CAPTION_OFFSET, chips, connectors}
}
