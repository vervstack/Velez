import {
    RunnerProvider,
    type Port,
    type RegisterContainerRequest,
    type RegisterContainerRequestRunnerPattern,
    type RunnerScope,
} from "@/app/api/velez"
import {parseConcurrent} from "@/processes/parseConcurrent.ts"
import {ResolvedLink} from "@/dialogs/CreateServiceDialog/processes/bindMounts.ts"

export type RegisterPattern = "generic" | "postgres" | "registry" | "runner"

export interface PgLogin {
    superuser: string
    password: string
}

export interface RegistryLogin {
    username: string
    password: string
}

export interface RunnerForm {
    provider?: RunnerProvider
    scope: RunnerScope
    target: string
    baseUrl: string
    labels: string
    dockerImage: string
    concurrent: string
    accessToken: string
    registrationToken: string
}

export interface RegisterContainerForm {
    containerId: string
    environment: string
    serviceName: string
    pattern: RegisterPattern
    pgLogin?: PgLogin
    registryLogin?: RegistryLogin
    runner?: RunnerForm
    links: ResolvedLink[]
    isClusterMode: boolean
    isKeepingPorts: boolean
    ports: Port[]
}

function isPgLoginIncomplete(form: RegisterContainerForm): boolean {
    if (form.pattern !== "postgres" || !form.pgLogin) return false
    return !form.pgLogin.superuser.trim() || !form.pgLogin.password
}

function isRegistryLoginIncomplete(form: RegisterContainerForm): boolean {
    if (form.pattern !== "registry" || !form.registryLogin) return false
    return !form.registryLogin.username.trim() || !form.registryLogin.password
}

function isGitlab(runner: RunnerForm): boolean {
    return runner.provider === RunnerProvider.GITLAB
}

function isRunnerInvalid(form: RegisterContainerForm): boolean {
    if (form.pattern !== "runner") return false
    const runner = form.runner
    if (!runner || !runner.provider || !runner.target.trim()) return true
    if (!isGitlab(runner) || !runner.concurrent.trim()) return false
    return parseConcurrent(runner.concurrent) === undefined
}

function buildRunnerPattern(runner: RunnerForm): RegisterContainerRequestRunnerPattern {
    const gitlab = isGitlab(runner)
    return {
        provider: runner.provider,
        scope: runner.scope,
        target: runner.target.trim(),
        labels: runner.labels.split(",").map((label) => label.trim()).filter((label) => label.length > 0),
        baseUrl: gitlab ? runner.baseUrl.trim() || undefined : undefined,
        dockerImage: gitlab ? runner.dockerImage.trim() || undefined : undefined,
        concurrent: gitlab ? parseConcurrent(runner.concurrent) : undefined,
        accessToken: runner.accessToken.trim() || undefined,
        registrationToken: runner.registrationToken.trim() || undefined,
    }
}

function isInvalid(form: RegisterContainerForm): boolean {
    if (!form.containerId || !form.serviceName.trim()) return true
    if (form.links.some((link) => !link.volumeName.trim())) return true
    if (isPgLoginIncomplete(form)) return true
    if (isRegistryLoginIncomplete(form)) return true
    if (isRunnerInvalid(form)) return true
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

    if (form.pattern === "runner") return form.runner ? {...base, runner: buildRunnerPattern(form.runner)} : null

    if (form.pattern === "registry") {
        const registry = form.registryLogin
            ? {username: form.registryLogin.username.trim(), password: form.registryLogin.password}
            : {}
        return {...base, registry}
    }

    const pg = form.pgLogin ? {superuser: form.pgLogin.superuser.trim(), password: form.pgLogin.password} : {}
    return {...base, pg}
}
