export interface BypassHostRow {
    id: number
    host: string
}

export interface ProxyPayload {
    proxyUrl: string
    proxyBypassHosts: string[]
}

export function newBypassHostRow(id: number, host: string = ""): BypassHostRow {
    return {id, host}
}

export function toBypassHostRows(hosts: string[]): BypassHostRow[] {
    return hosts.map((host, index) => newBypassHostRow(index + 1, host))
}

export function buildProxyPayload(proxyUrl: string, rows: BypassHostRow[]): ProxyPayload | null {
    const trimmedUrl = proxyUrl.trim()
    if (!trimmedUrl) return null

    const hosts = rows.map((row) => row.host.trim()).filter((host) => host !== "")
    return {proxyUrl: trimmedUrl, proxyBypassHosts: Array.from(new Set(hosts))}
}

export function buildRemoveProxyPayload(): ProxyPayload {
    return {proxyUrl: "", proxyBypassHosts: []}
}
