import type {ResourceAddress} from "@/model/service_page/ServicePageModel"

const DEFAULT_PING_TIMEOUT_MS = 3000

export function resourceAddressUrl(
    address: Pick<ResourceAddress, "host" | "port">,
    location: Pick<Location, "protocol" | "hostname">
): string {
    const host = address.host || location.hostname
    return `${location.protocol}//${host}:${address.port}`
}

// With mode "no-cors" a resolved fetch means the host answered; a rejection means it is unreachable.
export function pingAddress(url: string, timeoutMs = DEFAULT_PING_TIMEOUT_MS): Promise<boolean> {
    return fetch(url, {mode: "no-cors", cache: "no-store", signal: AbortSignal.timeout(timeoutMs)})
        .then(() => true)
        .catch(() => false)
}
