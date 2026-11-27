import type {
    CreateHaproxyInstanceRequest,
    HaproxyConnectionTarget,
} from "@/app/api/velez/haproxy_api.pb"

export const DEFAULT_PORT = 5432

export const networkLabels = {
    docker: "Shared docker network",
    public: "Public internet",
    tailscale: "Tailscale sidecar",
}

export type NetworkMode = keyof typeof networkLabels

export interface HaproxyWizardContext {
    name: string
    master: string
    masterLabel: string
    environment: string
    box: string
    dockerNetwork: string
    mode?: NetworkMode
    host: string
    targetPort: string
    tailscaleIp: string
    tailscalePort: string
    authKey: string
    loginServerUrl: string
    hostname: string
    isNetworkReady: boolean
    isPortExposed: boolean
    isPortReady: boolean
    port?: string
}

export const initialContext: HaproxyWizardContext = {
    name: "",
    master: "",
    masterLabel: "",
    environment: "",
    box: "",
    dockerNetwork: "",
    host: "",
    targetPort: String(DEFAULT_PORT),
    tailscaleIp: "",
    tailscalePort: String(DEFAULT_PORT),
    authKey: "",
    loginServerUrl: "",
    hostname: "",
    isNetworkReady: false,
    isPortExposed: false,
    isPortReady: false,
}

export function networkMode(context: HaproxyWizardContext): NetworkMode {
    return context.mode ?? (context.dockerNetwork ? "docker" : "public")
}

export function parsePort(port: string): number | undefined {
    const number = Number(port)
    return /^\d+$/.test(port) && number >= 1 && number <= 65535 ? number : undefined
}

export function formatEndpoint(host: string, port: string): string {
    return `${host.includes(":") ? `[${host}]` : host}:${port}`
}

export function normalizeHost(host: string): string {
    return host.trim().replace(/^\[([^\]]+)\]$/, "$1")
}

export function isIpAddress(host: string): boolean {
    if (/^\d{1,3}(\.\d{1,3}){3}$/.test(host)) {
        return host.split(".").every((part) => String(Number(part)) === part && Number(part) <= 255)
    }

    if (!/^[0-9a-f:.]+$/i.test(host) || !host.includes(":")) return false

    try {
        new URL(`http://[${host}]`)
        return true
    } catch {
        return false
    }
}

export function createRequest(context: HaproxyWizardContext): CreateHaproxyInstanceRequest {
    const mode = networkMode(context)
    let connectionTarget: HaproxyConnectionTarget

    switch (mode) {
        case "docker":
            connectionTarget = {dockerNetwork: context.dockerNetwork}
            break
        case "public":
            connectionTarget = {
                publicEndpoint: {host: normalizeHost(context.host), port: parsePort(context.targetPort)},
            }
            break
        case "tailscale":
            connectionTarget = {
                tailscaleEndpoint: {
                    host: normalizeHost(context.tailscaleIp), port: parsePort(context.tailscalePort),
                },
            }
            break
    }

    return {
        name: context.name.trim(),
        masterPgInstance: context.master,
        environment: mode === "docker" ? context.environment || undefined : undefined,
        box: context.box.trim() || undefined,
        connectionTarget,
        publishPort: context.isPortExposed,
        port: context.isPortExposed ? parsePort(context.port ?? "") : undefined,
    }
}
