function portOf(value: string): number | undefined {
    const port = Number(value.trim())

    if (!Number.isInteger(port) || port < 1 || port > 65535) {
        return undefined
    }

    return port
}

export function expandAvailablePorts(ranges: string[]): Set<number> {
    const availablePorts = new Set<number>()

    for (const range of ranges) {
        const parts = range.split(":")
        const start = portOf(parts[0])
        const end = parts.length === 1 ? start : portOf(parts[1])

        if (parts.length > 2 || start === undefined || end === undefined || start > end) {
            continue
        }

        for (let port = start; port <= end; port += 1) {
            availablePorts.add(port)
        }
    }

    return availablePorts
}

export function isPortAvailable(availablePorts: Set<number>, port: number): boolean {
    return availablePorts.has(port)
}
