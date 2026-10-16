import type {ResourceAddress} from "@/model/service_page/ServicePageModel"

export function resourceGoToUrl(
    addresses: ResourceAddress[],
    location: Pick<Location, "protocol" | "hostname">
): string | undefined {
    const address = pickGoToAddress(addresses, location.hostname)
    if (!address) {
        return undefined
    }
    const host = address.host || location.hostname
    return `${location.protocol}//${host}:${address.port}`
}

function pickGoToAddress(addresses: ResourceAddress[], pageHostname: string): ResourceAddress | undefined {
    if (addresses.length === 1) {
        return addresses[0]
    }
    return addresses.find(address => (address.host || pageHostname) === pageHostname)
}
