import type {RegistryInstance} from "@/app/api/velez"

export type RegistryInstanceStatus = "running" | "degraded" | "stopped"

export function mapRegistryInstanceStatus(status?: string): RegistryInstanceStatus {
    const s = (status ?? "").toLowerCase()
    if (s.includes("run")) return "running"
    if (s.includes("degrad") || s.includes("restart")) return "degraded"
    return "stopped"
}

export function formatRegistryInstanceCreatedAt(ts?: { seconds?: string | number }): string {
    const seconds = Number(ts?.seconds ?? 0)
    if (!seconds) return "-"
    return new Date(seconds * 1000).toLocaleDateString()
}

export function sortRegistryInstancesByName(instances: RegistryInstance[]): RegistryInstance[] {
    return [...instances].sort((a, b) => (a.name ?? "").localeCompare(b.name ?? ""))
}
