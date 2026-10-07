import {NetworkCapability} from "@/app/api/velez/network_api.pb"
import type {
    ConnectContainerRequest,
    CreateNetworkRequest,
    GetNetworkStatusResponse,
    Network,
} from "@/app/api/velez/network_api.pb"

export function shouldShowClusterBanner(status?: GetNetworkStatusResponse): boolean {
    return Boolean(status?.isClusterMode) && !status?.isVcnConnected
}

export function sortNetworks(networks: Network[]): Network[] {
    return [...networks].sort((a, b) => {
        if (Boolean(a.isManaged) !== Boolean(b.isManaged)) return a.isManaged ? -1 : 1
        return (a.name ?? "").localeCompare(b.name ?? "")
    })
}

function hasCapability(network: Network, capability: NetworkCapability): boolean {
    return (network.capabilities ?? []).includes(capability)
}

export function hasDeleteCapability(network: Network): boolean {
    return hasCapability(network, NetworkCapability.NETWORK_CAPABILITY_DELETE)
}

export function canDeleteNetwork(network: Network): boolean {
    return hasDeleteCapability(network) && (network.members ?? []).length === 0
}

export function canAttachNetwork(network: Network): boolean {
    return hasCapability(network, NetworkCapability.NETWORK_CAPABILITY_ATTACH)
}

export function getNetworkFlags(network: Network): string[] {
    const flags: string[] = []
    if (network.isInternal) flags.push("internal")
    if (network.isIccEnabled === false) flags.push("container isolation")
    if (network.subnet) flags.push(network.subnet)
    return flags
}

export function getNetworkKindLabel(network: Network): string {
    return network.isManaged ? "Velez" : "Docker (foreign)"
}

export function parseAliases(text: string): string[] {
    return text
        .split(",")
        .map((alias) => alias.trim())
        .filter((alias) => alias !== "")
}

export function getAttachableContainerNames(containerNames: string[], network: Network): string[] {
    const attached = new Set((network.members ?? []).map((member) => member.containerName))
    return containerNames.filter((name) => !attached.has(name)).sort((a, b) => a.localeCompare(b))
}

interface CreateNetworkInput {
    name: string
    environment: string
    isInternal: boolean
    isIsolatingContainers: boolean
}

export function buildCreateNetworkRequest(input: CreateNetworkInput): CreateNetworkRequest | null {
    const name = input.name.trim()
    if (!name) return null

    return {
        name,
        environment: input.environment,
        isInternal: input.isInternal,
        isIccEnabled: !input.isIsolatingContainers,
    }
}

interface ConnectContainerInput {
    network: Network
    containerName: string
    aliasesText: string
    environment: string
}

export function buildConnectContainerRequest(input: ConnectContainerInput): ConnectContainerRequest | null {
    if (!input.containerName || !input.network.id) return null

    return {
        networkId: input.network.id,
        containerName: input.containerName,
        aliases: parseAliases(input.aliasesText),
        environment: input.environment,
    }
}
