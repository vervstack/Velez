import type {CreateVcnNetworkRequest} from "@/app/api/velez/verv_closed_network.pb"
import {validateInstanceName} from "@/processes/mappings/instanceName.ts"

export interface VcnNetworkWizardContext {
    name: string
    source: "deploy" | "external"
    isExternalAvailable: boolean
    customPort: string
    customImage: string
    url: string
    token: string
}

export const initialContext: VcnNetworkWizardContext = {
    name: "",
    source: "deploy",
    isExternalAvailable: false,
    customPort: "",
    customImage: "",
    url: "",
    token: "",
}

export function canProceedFromName({name}: VcnNetworkWizardContext): boolean {
    return Boolean(name.trim()) && !validateInstanceName(name)
}

export function validateCustomPort(value: string): string | undefined {
    const port = value.trim()
    if (!port) return undefined

    if (!/^\d+$/.test(port) || Number(port) < 1 || Number(port) > 65535) {
        return "Enter a whole number between 1 and 65535."
    }
}

export function canProceedFromSource(context: VcnNetworkWizardContext): boolean {
    if (context.source === "external") {
        return context.isExternalAvailable && Boolean(context.url.trim() && context.token.trim())
    }

    return !validateCustomPort(context.customPort)
}

export function createRequest(context: VcnNetworkWizardContext): CreateVcnNetworkRequest {
    const name = context.name.trim()
    if (context.source === "external") {
        return {name, external: {url: context.url.trim(), token: context.token.trim()}}
    }

    const port = context.customPort.trim()
    return {
        name,
        deploy: {
            customPort: port ? Number(port) : undefined,
            customImage: context.customImage.trim() || undefined,
        },
    }
}
