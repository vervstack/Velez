export interface ContainerSearchQuery {
    mode: "substring" | "service"
    value: string
}

const SERVICE_PREFIX = /^service:\s*/i

export function parseContainerSearch(raw: string): ContainerSearchQuery {
    const trimmed = raw.trim()
    if (SERVICE_PREFIX.test(trimmed)) {
        return {mode: "service", value: trimmed.replace(SERVICE_PREFIX, "")}
    }
    return {mode: "substring", value: trimmed}
}
