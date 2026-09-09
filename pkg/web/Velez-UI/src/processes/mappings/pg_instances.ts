import type {PgInstance} from "@/app/api/velez"

export type PgInstanceStatus = "running" | "degraded" | "stopped"

export function mapPgInstanceStatus(status?: string): PgInstanceStatus {
    const s = (status ?? "").toLowerCase()
    if (s.includes("run")) return "running"
    if (s.includes("degrad") || s.includes("restart")) return "degraded"
    return "stopped"
}

export function formatPgInstanceCreatedAt(ts?: { seconds?: string | number }): string {
    const seconds = Number(ts?.seconds ?? 0)
    if (!seconds) return "-"
    return new Date(seconds * 1000).toLocaleDateString()
}

export function sortPgInstancesByName(instances: PgInstance[]): PgInstance[] {
    return [...instances].sort((a, b) => (a.name ?? "").localeCompare(b.name ?? ""))
}
