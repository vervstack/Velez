import {describe, expect, it} from "vitest"

import {TransferStatus} from "@/app/api/velez/transfer_api.pb"
import {
    canStartTransfer,
    hasActiveTransfer,
    isTransferActive,
    sortTransfersNewestFirst,
    statusLabel,
    transferPercent,
} from "@/processes/mappings/transfer.ts"

describe("transferPercent", () => {
    it("returns null when the total size is unknown", () => {
        const info = {status: TransferStatus.TRANSFER_STATUS_RUNNING, totalBytes: "0", doneBytes: "0"}

        expect(transferPercent(info)).toBeNull()
    })

    it("rounds done over total to a whole percent", () => {
        const info = {status: TransferStatus.TRANSFER_STATUS_RUNNING, totalBytes: "300", doneBytes: "100"}

        expect(transferPercent(info)).toBe(33)
    })

    it("caps the value at 100 when done exceeds total", () => {
        const info = {status: TransferStatus.TRANSFER_STATUS_RUNNING, totalBytes: "10", doneBytes: "50"}

        expect(transferPercent(info)).toBe(100)
    })

    it("returns 100 for a finished transfer even without sizes", () => {
        expect(transferPercent({status: TransferStatus.TRANSFER_STATUS_DONE})).toBe(100)
    })
})

describe("isTransferActive", () => {
    it("is true for pending and running only", () => {
        expect(isTransferActive(TransferStatus.TRANSFER_STATUS_PENDING)).toBe(true)
        expect(isTransferActive(TransferStatus.TRANSFER_STATUS_RUNNING)).toBe(true)
        expect(isTransferActive(TransferStatus.TRANSFER_STATUS_FAILED)).toBe(false)
        expect(isTransferActive(TransferStatus.TRANSFER_STATUS_DONE)).toBe(false)
        expect(isTransferActive(TransferStatus.TRANSFER_STATUS_CANCELLED)).toBe(false)
        expect(isTransferActive(undefined)).toBe(false)
    })
})

describe("hasActiveTransfer", () => {
    it("detects a running transfer in the list", () => {
        const transfers = [
            {status: TransferStatus.TRANSFER_STATUS_DONE},
            {status: TransferStatus.TRANSFER_STATUS_RUNNING},
        ]

        expect(hasActiveTransfer(transfers)).toBe(true)
        expect(hasActiveTransfer([])).toBe(false)
        expect(hasActiveTransfer(undefined)).toBe(false)
    })
})

describe("statusLabel", () => {
    it("maps every known status to a readable label", () => {
        expect(statusLabel(TransferStatus.TRANSFER_STATUS_CANCELLED)).toBe("Cancelled")
        expect(statusLabel(TransferStatus.TRANSFER_STATUS_UNSPECIFIED)).toBe("Unknown")
    })
})

describe("sortTransfersNewestFirst", () => {
    it("orders by creation time descending without mutating the input", () => {
        const older = {id: "a", createdAt: {seconds: "100"}}
        const newer = {id: "b", createdAt: {seconds: "200"}}
        const input = [older, newer]

        expect(sortTransfersNewestFirst(input).map((t) => t.id)).toEqual(["b", "a"])
        expect(input[0]).toBe(older)
    })
})

describe("canStartTransfer", () => {
    const base = {
        plan: {sidecars: [], consumers: [], unsupportedMounts: []},
        containerName: "app",
        address: "10.0.0.2:53890",
        token: "tok",
        isConsumersAcknowledged: false,
    }

    it("allows starting a clean plan", () => {
        expect(canStartTransfer(base)).toBe(true)
    })

    it("blocks without a plan or without address and token", () => {
        expect(canStartTransfer({...base, plan: undefined})).toBe(false)
        expect(canStartTransfer({...base, address: " "})).toBe(false)
        expect(canStartTransfer({...base, token: ""})).toBe(false)
    })

    it("blocks while consumers are not acknowledged", () => {
        const plan = {...base.plan, consumers: [{containerName: "web", networkName: "net"}]}

        expect(canStartTransfer({...base, plan})).toBe(false)
        expect(canStartTransfer({...base, plan, isConsumersAcknowledged: true})).toBe(true)
    })

    it("blocks on unsupported mounts even when acknowledged", () => {
        const plan = {...base.plan, unsupportedMounts: ["/etc/host"]}

        expect(canStartTransfer({...base, plan, isConsumersAcknowledged: true})).toBe(false)
    })
})
