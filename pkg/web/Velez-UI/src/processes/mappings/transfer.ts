import {TransferStatus} from "@/app/api/velez/transfer_api.pb"
import type {PlanTransferResponse, TransferInfo} from "@/app/api/velez/transfer_api.pb"

export type TransferTone = "pending" | "running" | "failed" | "done" | "cancelled"

const ACTIVE_STATUSES: TransferStatus[] = [
    TransferStatus.TRANSFER_STATUS_PENDING,
    TransferStatus.TRANSFER_STATUS_RUNNING,
]

const STATUS_LABELS: Record<string, string> = {
    [TransferStatus.TRANSFER_STATUS_PENDING]: "Pending",
    [TransferStatus.TRANSFER_STATUS_RUNNING]: "Running",
    [TransferStatus.TRANSFER_STATUS_FAILED]: "Failed",
    [TransferStatus.TRANSFER_STATUS_DONE]: "Done",
    [TransferStatus.TRANSFER_STATUS_CANCELLED]: "Cancelled",
}

const STATUS_TONES: Record<string, TransferTone> = {
    [TransferStatus.TRANSFER_STATUS_PENDING]: "pending",
    [TransferStatus.TRANSFER_STATUS_RUNNING]: "running",
    [TransferStatus.TRANSFER_STATUS_FAILED]: "failed",
    [TransferStatus.TRANSFER_STATUS_DONE]: "done",
    [TransferStatus.TRANSFER_STATUS_CANCELLED]: "cancelled",
}

const MAX_PERCENT = 100

export function isTransferActive(status?: TransferStatus): boolean {
    return status !== undefined && ACTIVE_STATUSES.includes(status)
}

export function isTransferFailed(status?: TransferStatus): boolean {
    return status === TransferStatus.TRANSFER_STATUS_FAILED
}

export function hasActiveTransfer(transfers?: TransferInfo[]): boolean {
    return (transfers ?? []).some((transfer) => isTransferActive(transfer.status))
}

export function statusLabel(status?: TransferStatus): string {
    return STATUS_LABELS[status ?? ""] ?? "Unknown"
}

export function statusTone(status?: TransferStatus): TransferTone {
    return STATUS_TONES[status ?? ""] ?? "pending"
}

export function transferPercent(info: TransferInfo): number | null {
    if (info.status === TransferStatus.TRANSFER_STATUS_DONE) return MAX_PERCENT

    const total = Number(info.totalBytes ?? 0)
    if (!Number.isFinite(total) || total <= 0) return null

    const done = Number(info.doneBytes ?? 0)
    const percent = Math.round((done / total) * MAX_PERCENT)
    return Math.min(MAX_PERCENT, Math.max(0, percent))
}

export function sortTransfersNewestFirst(transfers: TransferInfo[]): TransferInfo[] {
    return [...transfers].sort((a, b) => Number(b.createdAt?.seconds ?? 0) - Number(a.createdAt?.seconds ?? 0))
}

export function formatTransferTime(ts?: {seconds?: string | number}): string {
    const seconds = Number(ts?.seconds ?? 0)
    if (!seconds) return "-"
    return new Date(seconds * 1000).toLocaleString()
}

export interface StartTransferInput {
    plan?: PlanTransferResponse
    containerName: string
    address: string
    token: string
    isConsumersAcknowledged: boolean
}

export function canStartTransfer(input: StartTransferInput): boolean {
    if (!input.plan) return false
    if (!input.containerName || !input.address.trim() || !input.token.trim()) return false
    if ((input.plan.unsupportedMounts ?? []).length > 0) return false
    if ((input.plan.consumers ?? []).length > 0 && !input.isConsumersAcknowledged) return false
    return true
}
