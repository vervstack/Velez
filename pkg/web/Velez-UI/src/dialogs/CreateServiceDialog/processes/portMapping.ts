import type {DockerContainer, Port} from "@/app/api/velez"
import {PortProtocol} from "@/app/api/velez"

export interface PortMappingRow {
    containerPort: number
    protocol: PortProtocol
    currentHost: number
    newHost: string
}

export type RowOutcome = "kept" | "changed" | "unpublished"

const MAX_PORT = 65535

export function publishedPortsOf(container: DockerContainer): Port[] {
    return (container.ports ?? []).filter((port) => Boolean(port.exposedTo))
}

export function portMappingRowsOf(container: DockerContainer): PortMappingRow[] {
    return publishedPortsOf(container).map(function toRow(port) {
        const currentHost = port.exposedTo ?? 0
        return {
            containerPort: port.servicePortNumber ?? 0,
            protocol: port.protocol ?? PortProtocol.tcp,
            currentHost,
            newHost: String(currentHost),
        }
    })
}

export function rowOutcomeOf(row: PortMappingRow): RowOutcome {
    const value = row.newHost.trim()
    if (value === "") return "unpublished"
    return Number(value) === row.currentHost ? "kept" : "changed"
}

function parsePortNumber(raw: string): number | null {
    if (!/^\d+$/.test(raw)) return null

    const value = Number(raw)
    return value >= 1 && value <= MAX_PORT ? value : null
}

export function parsePortMappingRows(rows: PortMappingRow[]): Port[] | null {
    const ports: Port[] = []
    const seen = new Set<number>()

    for (const row of rows) {
        const raw = row.newHost.trim()
        if (raw === "") continue

        const exposedTo = parsePortNumber(raw)
        if (exposedTo === null || seen.has(exposedTo)) return null

        seen.add(exposedTo)
        ports.push({servicePortNumber: row.containerPort, exposedTo, protocol: row.protocol})
    }

    return ports
}

export function hasVolumes(container: DockerContainer): boolean {
    return (container.mounts?.length ?? 0) > 0
}

export function describeRowOutcome(row: PortMappingRow): string {
    const outcome = rowOutcomeOf(row)
    if (outcome === "kept") {
        return `Port ${row.currentHost} is kept: the service is stopped completely (not deleted) until you finish onboarding.`
    }
    if (outcome === "changed") {
        return `Port ${row.currentHost} -> ${row.newHost.trim()}: the old container is paused until you finish onboarding.`
    }
    return `Port ${row.containerPort} will not be published.`
}
