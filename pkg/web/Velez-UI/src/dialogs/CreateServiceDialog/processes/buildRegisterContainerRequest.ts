import type {Port, RegisterContainerRequest} from "@/app/api/velez"
import {ResolvedLink} from "@/dialogs/CreateServiceDialog/processes/bindMounts.ts"

export type RegisterPattern = "generic" | "postgres"

export interface PgLogin {
    superuser: string
    password: string
}

export interface RegisterContainerForm {
    containerId: string
    environment: string
    serviceName: string
    pattern: RegisterPattern
    pgLogin?: PgLogin
    links: ResolvedLink[]
    isClusterMode: boolean
    isKeepingPorts: boolean
    ports: Port[]
}

function isPgLoginIncomplete(form: RegisterContainerForm): boolean {
    if (form.pattern !== "postgres" || !form.pgLogin) return false
    return !form.pgLogin.superuser.trim() || !form.pgLogin.password
}

function isInvalid(form: RegisterContainerForm): boolean {
    if (!form.containerId || !form.serviceName.trim()) return true
    if (form.links.some((link) => !link.volumeName.trim())) return true
    if (isPgLoginIncomplete(form)) return true
    return !form.isClusterMode && form.isKeepingPorts && form.ports.length > 0
}

export function buildRegisterContainerRequest(form: RegisterContainerForm): RegisterContainerRequest | null {
    if (isInvalid(form)) return null

    const base = {
        containerId: form.containerId,
        environment: form.environment || undefined,
        serviceName: form.serviceName.trim(),
        bindMountLinks: form.links.map((link) => ({source: link.source, volumeName: link.volumeName.trim()})),
        ...(form.isClusterMode ? {} : {
            keepPortMapping: form.isKeepingPorts,
            ports: form.isKeepingPorts ? [] : form.ports,
        }),
    }

    if (form.pattern === "generic") return {...base, generic: {}}

    const pg = form.pgLogin ? {superuser: form.pgLogin.superuser.trim(), password: form.pgLogin.password} : {}
    return {...base, pg}
}
