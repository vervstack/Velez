import type {DockerContainer, Port} from "@/app/api/velez"
import {PortProtocol} from "@/app/api/velez"

export interface PortRow {
    containerPort: string
    hostPort: string
}

const MAX_PORT = 65535

export const EMPTY_PORT_ROW: PortRow = {containerPort: "", hostPort: ""}

export function publishedPortsOf(container: DockerContainer): Port[] {
    return (container.ports ?? []).filter((port) => Boolean(port.exposedTo))
}

export function describePort(port: Port): string {
    return `${port.exposedTo} → ${port.servicePortNumber}/${port.protocol ?? PortProtocol.tcp}`
}

function parsePortNumber(raw: string): number | null {
    const trimmed = raw.trim()
    if (!/^\d+$/.test(trimmed)) return null

    const value = Number(trimmed)
    return value >= 1 && value <= MAX_PORT ? value : null
}

function isBlankRow(row: PortRow): boolean {
    return row.containerPort.trim() === "" && row.hostPort.trim() === ""
}

export function parsePortRows(rows: PortRow[]): Port[] | null {
    const ports: Port[] = []

    for (const row of rows) {
        if (isBlankRow(row)) continue

        const servicePortNumber = parsePortNumber(row.containerPort)
        const exposedTo = parsePortNumber(row.hostPort)
        if (servicePortNumber === null || exposedTo === null) return null

        ports.push({servicePortNumber, exposedTo, protocol: PortProtocol.tcp})
    }

    return ports
}
