import type {PreviewRegisterContainerRequest} from "@/app/api/velez"
import {RegisterPattern} from "@/dialogs/CreateServiceDialog/processes/buildRegisterContainerRequest.ts"
import {
    isGenericOnly,
    RegisterRoles,
    roleIdsOf,
} from "@/dialogs/CreateServiceDialog/processes/registerRoles.ts"

export interface PreviewForm {
    containerId: string
    environment: string
    serviceName: string
    pattern: RegisterPattern
    roles: RegisterRoles
}

function patternOf(pattern: RegisterPattern): Pick<PreviewRegisterContainerRequest, "generic" | "pg" | "registry" | "runner"> {
    if (pattern === "postgres") return {pg: {}}
    if (pattern === "registry") return {registry: {}}
    if (pattern === "runner") return {runner: {}}
    return {generic: {}}
}

export function buildPreviewRequest(form: PreviewForm): PreviewRegisterContainerRequest | null {
    const serviceName = form.serviceName.trim()
    if (!form.containerId || !serviceName) return null

    const pattern = isGenericOnly(form.roles, form.containerId) ? "generic" : form.pattern

    return {
        containerId: form.containerId,
        environment: form.environment || undefined,
        serviceName,
        ...roleIdsOf(form.roles, form.containerId),
        ...patternOf(pattern),
    } as PreviewRegisterContainerRequest
}
