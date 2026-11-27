import type {CreateHaproxyInstanceRequest, PgBackend} from "@/app/api/velez/haproxy_api.pb"

export const DEFAULT_PORT = 5432

export interface HaproxyWizardContext {
    name: string
    box: string
    isPortExposed: boolean
    isPortReady: boolean
    port?: string
    isBootstrapping: boolean
    sidecarContainerId?: string
    backends: PgBackend[]
}

export const initialContext: HaproxyWizardContext = {
    name: "",
    box: "",
    isPortExposed: false,
    isPortReady: false,
    isBootstrapping: false,
    backends: [],
}

export function parsePort(port: string): number | undefined {
    const number = Number(port)
    return /^\d+$/.test(port) && number >= 1 && number <= 65535 ? number : undefined
}

export function createRequest(context: HaproxyWizardContext): CreateHaproxyInstanceRequest {
    return {
        name: context.name.trim(),
        box: context.box || undefined,
        port: context.isPortExposed ? parsePort(context.port ?? "") : undefined,
        publishPort: context.isPortExposed,
        sidecarContainerId: context.sidecarContainerId,
        backends: context.backends,
    }
}
